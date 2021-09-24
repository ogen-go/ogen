package gen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ogen-go/ogen"
)

func (g *Generator) generateMethods() error {
	for path, item := range g.spec.Paths {
		if item.Ref != "" {
			return fmt.Errorf("referenced paths are not supported")
		}

		if err := forEachOps(item, func(method string, op ogen.Operation) error {
			if err := g.generateMethod(path, strings.ToUpper(method), op); err != nil {
				return fmt.Errorf("%s: %w", method, err)
			}
			return nil
		}); err != nil {
			return fmt.Errorf("paths: %s: %w", path, err)
		}
	}

	sort.SliceStable(g.methods, func(i, j int) bool {
		return strings.Compare(g.methods[i].Path(), g.methods[j].Path()) < 0
	})

	return nil
}

func (g *Generator) generateMethod(path, method string, op ogen.Operation) error {
	params := make(map[ParameterLocation][]Parameter)
	for _, p := range op.Parameters {
		if p.Ref != "" {
			componentParam, found := g.componentsParameter(p.Ref)
			if !found {
				return fmt.Errorf("parameter by reference '%s' not found", p.Ref)
			}

			p = componentParam
		}

		param, err := g.parseParameter(p, path)
		if err != nil {
			return fmt.Errorf("parse parameter '%s': %w", p.Name, err)
		}

		ps := params[param.In]
		ps = append(ps, param)
		params[param.In] = ps
	}

	// Use path + method as unique identifier.
	methodName := pascal(path, strings.ToLower(method))
	parts, err := parsePath(path, params[LocationPath])
	if err != nil {
		return fmt.Errorf("parse path: %w", err)
	}

	m := &Method{
		Name:       methodName,
		PathParts:  parts,
		HTTPMethod: method,
		Parameters: params,
	}

	if op.RequestBody != nil {
		iface := g.createIface(methodName + "Requester")
		iface.addMethod(camel(methodName + "Requester"))

		rbody, err := g.generateRequestBody(methodName, op.RequestBody)
		if err != nil {
			return fmt.Errorf("requestBody: %w", err)
		}

		for _, schema := range rbody.Contents {
			schema.implement(iface)
		}

		m.RequestBody = rbody
		m.RequestType = iface.Name
	}

	if len(op.Responses) > 0 {
		iface := g.createIface(methodName + "Responder")
		iface.addMethod(camel(methodName + "Responder"))

		resp, err := g.generateResponses(methodName, op.Responses)
		if err != nil {
			return fmt.Errorf("responses: %w", err)
		}

		for _, resp := range resp.Responses {
			resp.implement(iface)
		}

		if def := resp.Default; def != nil {
			m.ResponseDefault = def
			def.implement(iface)
		}

		m.Responses = resp.Responses
		m.ResponseType = iface.Name
	}

	g.methods = append(g.methods, m)
	return nil
}

func parsePath(path string, params []Parameter) (parts []PathPart, err error) {
	lookup := func(name string) (Parameter, bool) {
		for _, p := range params {
			if p.SourceName == name {
				return p, true
			}
		}
		return Parameter{}, false
	}

	for _, s := range strings.Split(path, "/") {
		if len(s) == 0 {
			continue
		}
		if len(s) > 2 && s[0] == '{' && s[len(s)-1] == '}' {
			name := s[1 : len(s)-1]
			param, found := lookup(name)
			if !found {
				return nil, fmt.Errorf("parameter '%s' not found in path", name)
			}

			parts = append(parts, PathPart{Param: param})
			continue
		}

		parts = append(parts, PathPart{Raw: s})
	}
	return
}
