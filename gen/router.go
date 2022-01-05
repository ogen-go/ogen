package gen

import (
	"fmt"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/ir"
)

type Route struct {
	Method    string        // GET, POST, DELETE
	Operation *ir.Operation // getUserInfo
	Path      string        // /api/v1/user/{name}/info
}

// MethodRoute is route for one Method.
type MethodRoute struct {
	Method string
	Tree   RouteTree
}

// Add adds route to this tree.
func (m *MethodRoute) Add(r Route) error {
	if err := m.Tree.addRoute(r.Path, r.Operation); err != nil {
		return errors.Wrapf(err, "add route %s", r.Path)
	}
	return nil
}

// Router contains list of routes.
type Router struct {
	Methods []MethodRoute
}

// Add adds new route.
func (s *Router) Add(r Route) error {
	for _, m := range s.Methods {
		if m.Method == r.Method {
			if err := m.Add(r); err != nil {
				return errors.Wrapf(err, "update method %s", r.Method)
			}
			return nil
		}
	}

	m := MethodRoute{
		Method: r.Method,
		Tree:   RouteTree{},
	}
	if err := m.Add(r); err != nil {
		return errors.Wrapf(err, "update method %s", r.Method)
	}
	s.Methods = append(s.Methods, m)
	return nil
}

func printEdge(ident int, e *RouteNode) {
	identStr := strings.Repeat(" ", ident)
	p := e.Prefix()
	if param := e.Param(); param != nil {
		p = fmt.Sprintf(":%s", param.Spec.Name)
	}

	fmt.Printf("%s /%s", identStr, p)
	if e.IsLeaf() {
		fmt.Printf(" %s\n", e.Operation().Name)
		return
	}
	if op := e.Operation(); op != nil {
		fmt.Printf("/ %s\n", op.Name)
	} else {
		fmt.Printf("/\n")
	}
}

func (g *Generator) route() error {
	for _, op := range g.operations {
		if err := g.router.Add(Route{
			Method:    op.Spec.HTTPMethod,
			Path:      op.RawPath,
			Operation: op,
		}); err != nil {
			return errors.Wrapf(err, "add route %q", op.Name)
		}
	}

	if g.opt.VerboseRoute {
		for _, m := range g.router.Methods {
			fmt.Println(m.Method)
			m.Tree.Walk(func(level int, n *RouteNode) {
				printEdge(level*2, n)
			})
		}
	}

	return nil
}
