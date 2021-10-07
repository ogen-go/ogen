package gen

import (
	"sort"
	"strings"

	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/ast"
)

func (g *Generator) generateMethods() error {
	for path, item := range g.spec.Paths {
		if item.Ref != "" {
			return xerrors.New("referenced paths are not supported")
		}

		if g.opt.SpecificPath != "" {
			if g.opt.SpecificPath != path {
				continue
			}
		}

		if err := forEachOps(item, func(method string, op ogen.Operation) error {
			if err := g.generateMethod(path, strings.ToUpper(method), op); err != nil {
				return xerrors.Errorf("%s: %w", method, err)
			}
			return nil
		}); err != nil {
			return xerrors.Errorf("paths: %s: %w", path, err)
		}
	}

	sort.SliceStable(g.methods, func(i, j int) bool {
		return strings.Compare(g.methods[i].Path(), g.methods[j].Path()) < 0
	})

	return nil
}

func (g *Generator) generateMethod(path, method string, op ogen.Operation) (err error) {
	// Use path + method as unique identifier.
	methodName := pascal(path, strings.ToLower(method))
	if op.OperationID != "" {
		// Use operationId if present.
		methodName = pascal(op.OperationID)
	}

	m := &ast.Method{
		Name:       methodName,
		HTTPMethod: method,
	}

	m.Parameters, err = g.generateParams(path, op.Parameters)
	if err != nil {
		return xerrors.Errorf("parameters: %w", err)
	}

	m.PathParts, err = g.parsePath(path, m.PathParams())
	if err != nil {
		return xerrors.Errorf("parse path: %w", err)
	}

	if op.RequestBody != nil {
		iface := ast.Iface(methodName + "Request")
		iface.AddMethod(camel(methodName + "Request"))
		g.interfaces[iface.Name] = iface

		rbody, err := g.generateRequestBody(methodName, op.RequestBody)
		if err != nil {
			return xerrors.Errorf("requestBody: %w", err)
		}

		for _, schema := range rbody.Contents {
			schema.Implement(iface)
		}

		m.RequestBody = rbody
		m.RequestType = iface
	}

	if len(op.Responses) > 0 {
		iface := ast.Iface(methodName + "Response")
		iface.AddMethod(camel(methodName + "Response"))
		g.interfaces[iface.Name] = iface

		responses, err := g.generateResponses(methodName, op.Responses)
		if err != nil {
			return xerrors.Errorf("responses: %w", err)
		}

		for _, resp := range responses.StatusCode {
			resp.Implement(iface)
		}

		if def := responses.Default; def != nil {
			def.Implement(iface)
		}

		m.Responses = responses
		m.ResponseType = iface
	}

	g.methods = append(g.methods, m)
	return nil
}

func (g *Generator) parsePath(path string, params []*ast.Parameter) (parts []ast.PathPart, err error) {
	lookup := func(name string) (*ast.Parameter, bool) {
		for _, p := range params {
			if p.SourceName == name {
				return p, true
			}
		}
		return nil, false
	}

	for _, s := range strings.Split(path, "/") {
		if len(s) == 0 {
			continue
		}
		if len(s) > 2 && s[0] == '{' && s[len(s)-1] == '}' {
			name := s[1 : len(s)-1]
			param, found := lookup(name)
			if !found {
				return nil, xerrors.Errorf("path parameter '%s' not found in parameters", name)
			}

			parts = append(parts, ast.PathPart{Param: param})
			continue
		}

		parts = append(parts, ast.PathPart{Raw: s})
	}
	return
}
