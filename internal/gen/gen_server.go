package gen

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/ogen-go/ogen"
)

func parseParameter(param ogen.Parameter, path string) (*Parameter, error) {
	types := map[string]ParameterType{
		"query":  ParameterTypeQuery,
		"header": ParameterTypeHeader,
		"path":   ParameterTypePath,
		"cookie": ParameterCookie,
	}

	t, exists := types[strings.ToLower(param.In)]
	if !exists {
		return nil, fmt.Errorf("unsupported parameter type %s", param.In)
	}

	if t == ParameterTypePath {
		exists, err := regexp.MatchString(fmt.Sprintf("{%s}", param.Name), path)
		if err != nil {
			return nil, fmt.Errorf("match path param '%s': %w", param.Name, err)
		}

		if !exists {
			return nil, fmt.Errorf("param '%s' not found in path '%s'", param.Name, path)
		}
	}

	var allowArrayType bool
	if t == ParameterTypeHeader {
		allowArrayType = true
	}

	pType, err := parseSimpleType(param.Schema, parseSimpleTypeParams{
		AllowArrays: allowArrayType,
	})
	if err != nil {
		return nil, fmt.Errorf("parse type: %w", err)
	}

	return &Parameter{
		Name:       pascal(param.Name),
		SourceName: param.Name,
		Type:       pType,
		In:         t,
		Required:   param.Required,
	}, nil
}

func (g *Generator) generateServer() error {
	for path, item := range g.spec.Paths {
		if item.Ref != "" {
			return fmt.Errorf("reference objects in PathItem not supported yet")
		}

		if err := func() error {
			if err := g.generateOperation(path, http.MethodGet, item.Get); err != nil {
				return fmt.Errorf("method %s: %w", http.MethodGet, err)
			}
			if err := g.generateOperation(path, http.MethodPut, item.Put); err != nil {
				return fmt.Errorf("method %s: %w", http.MethodPut, err)
			}
			if err := g.generateOperation(path, http.MethodPost, item.Post); err != nil {
				return fmt.Errorf("method %s: %w", http.MethodPost, err)
			}
			if err := g.generateOperation(path, http.MethodDelete, item.Delete); err != nil {
				return fmt.Errorf("method %s: %w", http.MethodDelete, err)
			}
			if err := g.generateOperation(path, http.MethodOptions, item.Options); err != nil {
				return fmt.Errorf("method %s: %w", http.MethodOptions, err)
			}
			if err := g.generateOperation(path, http.MethodHead, item.Head); err != nil {
				return fmt.Errorf("method %s: %w", http.MethodHead, err)
			}
			if err := g.generateOperation(path, http.MethodPatch, item.Patch); err != nil {
				return fmt.Errorf("method %s: %w", http.MethodPatch, err)
			}
			if err := g.generateOperation(path, http.MethodTrace, item.Trace); err != nil {
				return fmt.Errorf("method %s: %w", http.MethodTrace, err)
			}
			return nil
		}(); err != nil {
			return fmt.Errorf("path %s: %w", path, err)
		}
	}

	sort.SliceStable(g.server.Methods, func(i, j int) bool {
		return strings.Compare(g.server.Methods[i].Path, g.server.Methods[j].Path) < 0
	})

	return nil
}

func (g *Generator) generateOperation(path, httpMethod string, op *ogen.Operation) error {
	if op == nil {
		return nil
	}

	if httpMethod != http.MethodPost &&
		httpMethod != http.MethodPut &&
		httpMethod != http.MethodPatch &&
		op.RequestBody != nil {
		return fmt.Errorf("requestBody is not supported for this http method")
	}

	method := serverMethodDef{
		Name:        toFirstUpper(op.OperationID),
		OperationID: op.OperationID,
		Path:        path,
		HTTPMethod:  strings.ToUpper(httpMethod),
	}

	if body := op.RequestBody; body != nil {
		// Try to get request body from components section.
		if body.Ref != "" {
			rbody, found := g.componentsRequestBody(body.Ref)
			if !found {
				return fmt.Errorf("parse requestBody: ref '%s' not found", body.Ref)
			}

			body = &rbody
		}

		method.RequestBodyRequired = body.Required
		if len(body.Content) > 1 {
			return fmt.Errorf("parse requestBody: multiple contents not supported yet")
		}

		for contentType, media := range body.Content {
			name := g.schemaComponentGotype(media.Schema.Ref)
			if name == "" {
				return fmt.Errorf("parse requestBody: %s: ref %s not found", contentType, media.Schema.Ref)
			}

			method.RequestType = name
		}
	}

	for status, resp := range op.Responses {
		if status != "200" {
			return fmt.Errorf("parse responses: unsupported status code: %s", status)
		}

		// Try to get response from components section.
		if resp.Ref != "" {
			cresp, found := g.componentsResponse(resp.Ref)
			if !found {
				return fmt.Errorf("parse response: ref '%s' not found", resp.Ref)
			}

			resp = cresp
		}

		if len(resp.Content) > 1 {
			return fmt.Errorf("parse response: %s: multiple contents not supported yet", status)
		}

		for contentType, media := range resp.Content {
			name := g.schemaComponentGotype(media.Schema.Ref)
			if name == "" {
				return fmt.Errorf("parse response: %s: %s: ref %s not found", status, contentType, media.Schema.Ref)
			}

			method.ResponseType = name
		}
	}

	if len(op.Parameters) != 0 {
		method.Parameters = make(map[ParameterType][]Parameter)
	}

	for _, param := range op.Parameters {
		// Try to get param from components section.
		if param.Ref != "" {
			cparam, found := g.componentsParameter(param.Ref)
			if !found {
				return fmt.Errorf("parse parameters: ref '%s' not found", param.Ref)
			}

			param = cparam
		}

		parameter, err := parseParameter(param, path)
		if err != nil {
			return fmt.Errorf("parse method %s parameter %s: %w", op.OperationID, param.Name, err)
		}

		if _, exists := method.Parameters[parameter.In]; !exists {
			method.Parameters[parameter.In] = []Parameter{}
		}

		method.Parameters[parameter.In] = append(method.Parameters[parameter.In], *parameter)
	}

	g.server.Methods = append(g.server.Methods, method)
	return nil
}
