package gen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/ir"
)

// Router state for routing path to handlers.
type Router struct {
	Methods []RouterMethod

	routes []Route
}

type RouterMethod struct {
	Method string
	Edges  []*RouteEge
}

func (m RouterMethod) Root() *RouteEge {
	return &RouteEge{
		ID:   0,
		Next: m.Edges,
	}
}

func printEdge(ident int, e *RouteEge) {
	prefix := strings.Repeat(" ", ident)
	p := e.Path
	if e.Param != nil {
		p = fmt.Sprintf(":%s", e.Param.Name)
	}

	fmt.Printf("%s[%02d] /%s", prefix, e.ID, p)
	if len(e.Next) == 0 {
		fmt.Printf(" %s\n", e.Operation)
		return
	}
	if e.Operation != "" {
		fmt.Printf("/ %s\n", e.Operation)
	} else {
		fmt.Printf("/\n")
	}
	for _, next := range e.Next {
		printEdge(ident+2, next)
	}
}

type RouteEge struct {
	ID        int
	Path      string // path part
	Operation string // operation name if not blank
	Param     *ir.Parameter
	Next      []*RouteEge
}

type Route struct {
	Method    string         // GET, POST, DELETE
	Operation string         // getUserInfo
	Path      []*ir.PathPart // /api/v1/user/{name}/info
}

func (r *Router) Register(route Route) {
	r.routes = append(r.routes, route)
}

func (r *Router) Graph() error {
	methods := make(map[string][]int)
	for i, route := range r.routes {
		methods[route.Method] = append(methods[route.Method], i)
	}
	var allMethods []string
	for k := range methods {
		allMethods = append(allMethods, k)
	}
	sort.Strings(allMethods)
	for _, method := range allMethods {
		m := RouterMethod{
			Method: method,
		}
		var id int
		for _, i := range methods[method] {
			var edge *RouteEge

			route := r.routes[i]

			if len(route.Path) == 0 {
				id++
				m.Edges = append(m.Edges, &RouteEge{
					ID:        id,
					Operation: route.Operation,
				})

				continue
			}

		Path:
			for _, elem := range route.Path {
				raw := strings.ReplaceAll(elem.Raw, "/", "")

				edges := m.Edges
				if edge != nil {
					edges = edge.Next
				}
				for _, next := range edges {
					if next.Path == raw {
						edge = next
						continue Path
					}
				}
				if edge == nil {
					// Create new root edge.
					id++
					edge = &RouteEge{
						ID:    id,
						Path:  raw,
						Param: elem.Param,
					}
					m.Edges = append(m.Edges, edge)
					continue Path
				}

				id++
				nextEdge := &RouteEge{
					ID:    id,
					Path:  raw,
					Param: elem.Param,
				}
				edge.Next = append(edge.Next, nextEdge)
				edge = nextEdge
			}

			edge.Operation = route.Operation
		}
		r.Methods = append(r.Methods, m)
	}

	return nil
}

func (g *Generator) route() error {
	for _, op := range g.operations {
		g.router.Register(Route{
			Method:    op.Spec.HTTPMethod,
			Path:      op.PathParts,
			Operation: op.Name,
		})
	}
	if err := g.router.Graph(); err != nil {
		return errors.Wrap(err, "graph")
	}

	for _, m := range g.router.Methods {
		fmt.Println(m.Method)
		for _, e := range m.Edges {
			printEdge(2, e)
		}
	}

	return nil
}
