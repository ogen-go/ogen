package gen

import (
	"net/textproto"
	"slices"
	"sort"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/gen/ir"
)

// idSeq is a monotonically increasing id sequence.
type idSeq struct {
	id  int
	seq *int
}

func (s *idSeq) next() idSeq {
	if s.seq == nil {
		s.seq = new(int)
	}
	*s.seq++
	return idSeq{
		id:  *s.seq,
		seq: s.seq,
	}
}

type nodes []*RouteNode

func (e nodes) Len() int {
	return len(e)
}

func (e nodes) Less(i, j int) bool {
	return e[i].head < e[j].head
}

func (e nodes) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e nodes) Sort() {
	sort.Sort(e)
}

// RouteNode is node of Radix tree of routes.
type RouteNode struct {
	idSeq idSeq

	prefix string
	head   byte
	child  nodes

	paramName string
	param     *ir.Parameter // nil-able

	routes Routes
}

// ID returns node identifier.
func (n *RouteNode) ID() int {
	return n.idSeq.id
}

// AddRoute adds new method route to node.
func (n *RouteNode) AddRoute(nr Route) error {
	return n.routes.AddRoute(nr)
}

// Prefix returns common prefix.
func (n *RouteNode) Prefix() string {
	return n.prefix
}

// Head returns first byte of prefix.
func (n *RouteNode) Head() byte {
	return n.head
}

// IsStatic whether node is not a parameter node.
func (n *RouteNode) IsStatic() bool {
	return n.param == nil
}

// IsLeaf whether node has no children.
func (n *RouteNode) IsLeaf() bool {
	return n.child.Len() == 0
}

// IsParam whether node is a parameter node.
func (n *RouteNode) IsParam() bool {
	return n.param != nil
}

// Children returns child nodes.
func (n *RouteNode) Children() []*RouteNode {
	return n.child
}

// StaticChildren returns slice of child static nodes.
func (n *RouteNode) StaticChildren() (r []*RouteNode) {
	for _, c := range n.child {
		if c.IsStatic() {
			r = append(r, c)
		}
	}
	return r
}

// ParamChildren returns slice of child parameter nodes.
func (n *RouteNode) ParamChildren() (r []*RouteNode) {
	for _, c := range n.child {
		if c.IsParam() {
			r = append(r, c)
		}
	}
	return r
}

// Tails returns heads of child nodes.
//
// Used for matching end of parameter node between two static.
func (n *RouteNode) Tails() (r []byte) {
	for _, c := range n.child {
		if !c.IsParam() {
			r = append(r, c.head)
		}
	}
	return r
}

// ParamName returns parameter name, if any.
func (n *RouteNode) ParamName() string {
	return n.paramName
}

// Param returns associated parameter, if any.
//
// May be nil.
func (n *RouteNode) Param() *ir.Parameter {
	return n.param
}

// AllowedMethods returns list of allowed methods.
func (n *RouteNode) AllowedMethods() string {
	var s strings.Builder
	for i, route := range n.routes {
		if i != 0 {
			s.WriteByte(',')
		}
		s.WriteString(route.Method)
	}
	return s.String()
}

// WithAllowedHeaders reports whether any route
// in this node accepts any headers.
func (n *RouteNode) WithAllowedHeaders() bool {
	for _, route := range n.routes {
		if route.Operation.HasHeaderParams() {
			return true
		}
		if route.Operation.Request != nil {
			return true
		}
		for _, sec := range route.Operation.Security.Securities {
			if sec.Format == ir.APIKeySecurityFormat {
				if sec.Kind == ir.HeaderSecurity {
					return true
				}
			} else {
				return true
			}
		}
	}
	return false
}

// AllowedHeaders returns HTTP method and allowed headers pairs.
// Allowed headers are formatted as a comma-separated list of headers.
func (n *RouteNode) AllowedHeaders() [][2]string {
	var kvs [][2]string
	for _, route := range n.routes {
		var headers []string

		headerParams := route.Operation.HeaderParams()
		for _, param := range headerParams {
			headers = append(headers, textproto.CanonicalMIMEHeaderKey(param.Spec.Name))
		}

		if route.Operation.Request != nil {
			headers = append(headers, "Content-Type")
		}

		securities := route.Operation.Security.Securities
		for _, sec := range securities {
			if sec.Format == ir.APIKeySecurityFormat {
				if sec.Kind == ir.HeaderSecurity {
					headers = append(headers, textproto.CanonicalMIMEHeaderKey(sec.ParameterName))
				}
			} else {
				headers = append(headers, "Authorization")
			}
		}

		if len(headers) == 0 {
			continue
		}

		slices.Sort(headers)
		headers = slices.Compact(headers)

		kvs = append(kvs, [2]string{
			route.Method,
			strings.Join(headers, ","),
		})
	}
	return kvs
}

// PostContentTypes returns comma-separated list of content types
// accepted by a route with POST method.
func (n *RouteNode) PostContentTypes() string {
	return n.methodContentTypes("POST")
}

// PatchContentTypes returns comma-separated list of content types
// accepted by a route with PATCH method.
func (n *RouteNode) PatchContentTypes() string {
	return n.methodContentTypes("PATCH")
}

// methodContentTypes returns comma-separated list of content types
// accepted by a route with the specified HTTP method.
func (n *RouteNode) methodContentTypes(method string) string {
	var types []string
	for _, route := range n.routes {
		if route.Method != method {
			continue
		}
		if route.Operation.Request == nil {
			return ""
		}
		for typ := range route.Operation.Request.Contents {
			types = append(types, string(typ))
		}
		break
	}
	slices.Sort(types)
	return strings.Join(types, ",")
}

// Routes returns list of associated Route.
func (n *RouteNode) Routes() []Route {
	return n.routes
}

func nextPathPart(s string) (hasParam bool, paramStart, paramEnd int, _ error) {
	paramStart = strings.IndexByte(s, '{')
	if paramStart < 0 {
		return false, 0, 0, nil
	}

	paramEnd = strings.IndexByte(s, '}')
	if paramEnd < 0 || paramEnd < paramStart {
		return false, paramStart, paramEnd, errors.Errorf("unclosed '{' at %d", paramStart)
	}
	// Need to match parameter part including both brackets.
	paramEnd++
	return true, paramStart, paramEnd, nil
}

func (n *RouteNode) addChild(path string, op *ir.Operation, ch *RouteNode) (r *RouteNode, _ error) {
	r = ch

	hasParam, start, end, err := nextPathPart(path)
	if err != nil {
		return nil, errors.Errorf("parse %q", path)
	}

	if hasParam {
		paramName := path[start+1 : end-1]
		p, ok := findParam(op, paramName)
		if !ok {
			return nil, errors.Errorf("unknown parameter %q", paramName)
		}

		if start == 0 { // Route starts with a param.
			ch.paramName = paramName
			ch.param = p

			// Handle tail of path.
			if len(path[end:]) > 0 {
				path = path[end:]
				n, err := ch.addChild(path, op, &RouteNode{
					idSeq:  n.idSeq.next(),
					prefix: path,
					head:   path[0],
				})
				if err != nil {
					return nil, err
				}
				r = n
			}
		} else { // Route contains param.
			// Set prefix to static part of path.
			ch.prefix = path[:start]
			// Get parameterized part.
			path = path[start:]
			// Add parameterized child node.
			n, err := ch.addChild(path, op, &RouteNode{
				idSeq:     n.idSeq.next(),
				head:      path[0],
				paramName: paramName,
				param:     p,
			})
			if err != nil {
				return nil, err
			}
			r = n
		}
	}

	n.child = append(n.child, ch)
	n.child.Sort()
	return r, nil
}

func (n *RouteNode) childIdx(head byte) (int, bool) {
	for i := range n.child {
		if n.child[i].head == head {
			return i, true
		}
	}
	return 0, false
}

func (n *RouteNode) replaceChild(head byte, child *RouteNode) {
	idx, _ := n.childIdx(head)
	n.child[idx] = child
}

func (n *RouteNode) getChild(head byte) *RouteNode {
	idx, ok := n.childIdx(head)
	if ok && n.child[idx].head == head {
		return n.child[idx]
	}
	return nil
}

func (n *RouteNode) walk(level int, cb func(level int, n *RouteNode)) {
	if n == nil {
		return
	}
	cb(level, n)
	for _, ch := range n.child {
		ch.walk(level+1, cb)
	}
}
