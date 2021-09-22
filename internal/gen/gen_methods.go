package gen

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/ogen-go/ogen"
)

func (g *Generator) generateMethods() error {
	for path, item := range g.spec.Paths {
		if item.Ref != "" {
			return fmt.Errorf("referenced paths are not supported")
		}

		if err := func() error {
			if err := g.generateMethod(path, http.MethodGet, item.Get); err != nil {
				return fmt.Errorf("method %s: %w", http.MethodGet, err)
			}
			if err := g.generateMethod(path, http.MethodPut, item.Put); err != nil {
				return fmt.Errorf("method %s: %w", http.MethodPut, err)
			}
			if err := g.generateMethod(path, http.MethodPost, item.Post); err != nil {
				return fmt.Errorf("method %s: %w", http.MethodPost, err)
			}
			if err := g.generateMethod(path, http.MethodDelete, item.Delete); err != nil {
				return fmt.Errorf("method %s: %w", http.MethodDelete, err)
			}
			if err := g.generateMethod(path, http.MethodOptions, item.Options); err != nil {
				return fmt.Errorf("method %s: %w", http.MethodOptions, err)
			}
			if err := g.generateMethod(path, http.MethodHead, item.Head); err != nil {
				return fmt.Errorf("method %s: %w", http.MethodHead, err)
			}
			if err := g.generateMethod(path, http.MethodPatch, item.Patch); err != nil {
				return fmt.Errorf("method %s: %w", http.MethodPatch, err)
			}
			if err := g.generateMethod(path, http.MethodTrace, item.Trace); err != nil {
				return fmt.Errorf("method %s: %w", http.MethodTrace, err)
			}
			return nil
		}(); err != nil {
			return fmt.Errorf("path %s: %w", path, err)
		}
	}

	sort.SliceStable(g.methods, func(i, j int) bool {
		return strings.Compare(g.methods[i].Path(), g.methods[j].Path()) < 0
	})

	return nil
}

func (g *Generator) generateMethod(path, method string, op *ogen.Operation) error {
	if op == nil {
		return nil
	}

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
	methodName := strings.ReplaceAll(path, "/", "_")
	methodName = strings.ReplaceAll(methodName, "{", "")
	methodName = strings.ReplaceAll(methodName, "}", "")
	methodName += "_" + strings.ToLower(method)
	methodName = pascal(methodName)

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
		iface := g.createIface(methodName + "Request")
		iface.addMethod(camel(methodName + "Request"))

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
		iface := g.createIface(methodName + "Response")
		iface.addMethod(camel(methodName + "Response"))

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
