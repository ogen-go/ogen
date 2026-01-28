package gen

import (
	"net/textproto"
	"slices"
	"sort"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/internal/xmaps"
	"github.com/ogen-go/ogen/openapi"
)

// Route describes route.
type Route struct {
	Method    string        // GET, POST, DELETE
	Path      string        // /api/v1/user/{name}/info
	Operation *ir.Operation // getUserInfo
}

// Routes is list of routes.
type Routes []Route

// Len implements sort.Interface.
func (n Routes) Len() int {
	return len(n)
}

// Less implements sort.Interface.
func (n Routes) Less(i, j int) bool {
	return n[i].Method < n[j].Method
}

// Swap implements sort.Interface.
func (n Routes) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

// AddRoute adds new route. If the route is already added, it returns error.
func (n *Routes) AddRoute(nr Route) error {
	if slices.ContainsFunc(*n, func(r Route) bool { return strings.EqualFold(r.Method, nr.Method) }) {
		return errors.Errorf("duplicate method %q", nr.Method)
	}
	*n = append(*n, nr)
	// Keep routes sorted by method.
	sort.Sort(n)
	return nil
}

// Router contains list of routes.
type Router struct {
	Tree RouteTree
	// MaxParametersCount is maximum number of path parameters in one operation.
	MaxParametersCount int
}

// Add adds new route.
func (s *Router) Add(r Route) error {
	return s.Tree.addRoute(r)
}

// NodesWithAllowedHeaders returns route nodes that
// contain routes that accept headers.
func (s Router) NodesWithAllowedHeaders() []*RouteNode {
	var nodes []*RouteNode
	s.Tree.Walk(func(level int, n *RouteNode) {
		if n.WithAllowedHeaders() {
			nodes = append(nodes, n)
		}
	})
	return nodes
}

// WebhookRoute is a webhook route.
type WebhookRoute struct {
	Method    string
	Operation *ir.Operation
}

// WebhookRoutes is a list of webhook methods.
type WebhookRoutes struct {
	Routes []WebhookRoute
	idSeq  idSeq
}

// ID returns list identifier.
func (r WebhookRoutes) ID() int {
	return r.idSeq.id
}

// Add adds new operation to the route.
func (r *WebhookRoutes) Add(nr WebhookRoute) error {
	if slices.ContainsFunc(r.Routes, func(r WebhookRoute) bool {
		return strings.EqualFold(r.Method, nr.Method)
	}) {
		return errors.Errorf("duplicate method %q", nr.Method)
	}
	r.Routes = append(r.Routes, nr)
	slices.SortStableFunc(r.Routes, func(a, b WebhookRoute) int {
		return strings.Compare(a.Method, b.Method)
	})
	return nil
}

// AllowedMethods returns comma-separated list of allowed methods.
func (r WebhookRoutes) AllowedMethods() string {
	var s strings.Builder
	for i, route := range r.Routes {
		if i != 0 {
			s.WriteByte(',')
		}
		s.WriteString(route.Method)
	}
	return s.String()
}

// WithAllowedHeaders reports whether any route
// in this list accepts any headers.
func (r WebhookRoutes) WithAllowedHeaders() bool {
	for _, route := range r.Routes {
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
func (r WebhookRoutes) AllowedHeaders() [][2]string {
	var kvs [][2]string
	for _, route := range r.Routes {
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
func (r WebhookRoutes) PostContentTypes() string {
	return r.methodContentTypes("POST")
}

// PatchContentTypes returns comma-separated list of content types
// accepted by a route with PATCH method.
func (r WebhookRoutes) PatchContentTypes() string {
	return r.methodContentTypes("PATCH")
}

// methodContentTypes returns comma-separated list of content types
// accepted by a route with the specified HTTP method.
func (r WebhookRoutes) methodContentTypes(method string) string {
	var types []string
	for _, route := range r.Routes {
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

// WebhookRouter contains routing information for webhooks.
type WebhookRouter struct {
	Webhooks map[string]WebhookRoutes
	idSeq    idSeq
}

// WebhooksWithAllowedHeaders returns lists that
// contain routes that accept any headers.
func (r WebhookRouter) WebhooksWithAllowedHeaders() []WebhookRoutes {
	var hooks []WebhookRoutes
	for _, key := range xmaps.SortedKeys(r.Webhooks) {
		route := r.Webhooks[key]
		if route.WithAllowedHeaders() {
			hooks = append(hooks, route)
		}
	}
	return hooks
}

// Add adds new route.
func (r *WebhookRouter) Add(name string, nr WebhookRoute) error {
	if r.Webhooks == nil {
		r.Webhooks = map[string]WebhookRoutes{}
	}
	route, ok := r.Webhooks[name]
	if !ok {
		route = WebhookRoutes{idSeq: r.idSeq.next()}
	}
	if err := route.Add(nr); err != nil {
		return errors.Wrapf(err, "webhook %q", name)
	}
	r.Webhooks[name] = route
	return nil
}

func checkRoutePath(p openapi.Path) error {
	for i, part := range p {
		if i == 0 {
			continue
		}
		// Cond: i > 0
		current := part.Param
		prev := p[i-1].Param
		if prev != nil && current != nil {
			return errors.Errorf(
				"can't handle two parameters in a row (%q and %q)",
				prev.Name, current.Name,
			)
		}
	}
	return nil
}

func (g *Generator) route() error {
	var maxParametersCount int
	for _, op := range g.operations {
		path := op.Spec.Path

		if err := func() error {
			if err := checkRoutePath(path); err != nil {
				return err
			}
			return g.router.Add(Route{
				Method:    strings.ToUpper(op.Spec.HTTPMethod),
				Path:      path.String(),
				Operation: op,
			})
		}(); err != nil {
			return errors.Wrapf(err, "add route %q (op %q)", path, op.Name)
		}

		if count := op.PathParamsCount(); maxParametersCount < count {
			maxParametersCount = count
		}
	}
	g.router.MaxParametersCount = maxParametersCount
	for _, op := range g.webhooks {
		webhookName := op.WebhookInfo.Name
		nr := WebhookRoute{
			Method:    strings.ToUpper(op.Spec.HTTPMethod),
			Operation: op,
		}
		if err := g.webhookRouter.Add(webhookName, nr); err != nil {
			return errors.Wrap(err, "add route")
		}
	}
	return nil
}
