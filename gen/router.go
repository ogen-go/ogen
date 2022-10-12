package gen

import (
	"sort"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/internal/xslices"
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
	if xslices.ContainsFunc(*n, func(r Route) bool { return strings.EqualFold(r.Method, nr.Method) }) {
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
	return nil
}
