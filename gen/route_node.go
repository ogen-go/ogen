package gen

import (
	"sort"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/gen/ir"
)

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
	prefix string
	head   byte
	child  nodes

	paramName string
	param     *ir.Parameter // nil-able

	routes Routes
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

// Routes returns list of associated MethodRoute.
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
