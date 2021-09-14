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
		return strings.Compare(g.methods[i].Path, g.methods[j].Path) < 0
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

		param, err := parseParameter(p, path)
		if err != nil {
			return fmt.Errorf("parse parameter '%s': %w", p.Name, err)
		}

		ps := params[param.In]
		ps = append(ps, param)
		params[param.In] = ps
	}

	// Use path + method as unique identifier.
	name := strings.ReplaceAll(path, "/", "_")
	name = strings.ReplaceAll(name, "{", "")
	name = strings.ReplaceAll(name, "}", "")
	name += "_" + strings.ToLower(method)
	name = pascal(name)
	m := &Method{
		Name:       name,
		Path:       path,
		HTTPMethod: method,
		Parameters: params,
	}

	if op.RequestBody != nil {
		rbody, err := g.generateRequestBody(name, op.RequestBody)
		if err != nil {
			return fmt.Errorf("requestBody: %w", err)
		}

		for _, schema := range rbody.Contents {
			g.implementRequest(schema, m)
		}

		m.RequestBody = rbody
		m.RequestType = name + "Request"
	}

	g.methods = append(g.methods, m)
	return nil
}
