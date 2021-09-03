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
	}, nil
}

func (g *Generator) generateServer() error {
	for path, item := range g.spec.Paths {
		if item.Ref != "" {
			return fmt.Errorf("reference objects in PathItem not supported yet")
		}

		if err := g.generateOperation(path, http.MethodGet, item.Get); err != nil {
			return fmt.Errorf("generate op: %w", err)
		}
		if err := g.generateOperation(path, http.MethodPut, item.Put); err != nil {
			return fmt.Errorf("generate op: %w", err)
		}
		if err := g.generateOperation(path, http.MethodPost, item.Post); err != nil {
			return fmt.Errorf("generate op: %w", err)
		}
		if err := g.generateOperation(path, http.MethodDelete, item.Delete); err != nil {
			return fmt.Errorf("generate op: %w", err)
		}
		if err := g.generateOperation(path, http.MethodOptions, item.Options); err != nil {
			return fmt.Errorf("generate op: %w", err)
		}
		if err := g.generateOperation(path, http.MethodHead, item.Head); err != nil {
			return fmt.Errorf("generate op: %w", err)
		}
		if err := g.generateOperation(path, http.MethodPatch, item.Patch); err != nil {
			return fmt.Errorf("generate op: %w", err)
		}
		if err := g.generateOperation(path, http.MethodTrace, item.Trace); err != nil {
			return fmt.Errorf("generate op: %w", err)
		}
	}

	sort.SliceStable(g.server.Methods, func(i, j int) bool {
		return strings.Compare(g.server.Methods[i].Path, g.server.Methods[j].Path) < 0 ||
			strings.Compare(g.server.Methods[i].HTTPMethod, g.server.Methods[j].HTTPMethod) < 0
	})

	return nil
}

func (g *Generator) generateOperation(path, httpMethod string, op *ogen.Operation) error {
	if op == nil {
		return nil
	}

	method := serverMethodDef{
		Name:        toFirstUpper(op.OperationID),
		OperationID: op.OperationID,
		Path:        path,
		HTTPMethod:  strings.ToUpper(httpMethod),
	}

	for _, content := range op.RequestBody.Content {
		name := g.componentByRef(content.Schema.Ref)
		if name == "" {
			return fmt.Errorf("ref %s not found", content.Schema.Ref)
		}

		method.RequestType = name
	}

	for status, resp := range op.Responses {
		if status != "200" {
			return fmt.Errorf("unsupported response status code: %s", status)
		}

		for _, content := range resp.Content {
			name := g.componentByRef(content.Schema.Ref)
			if name == "" {
				return fmt.Errorf("ref %s not found", content.Schema.Ref)
			}

			method.ResponseType = name
		}
	}

	if len(op.Parameters) != 0 {
		method.Parameters = make(map[ParameterType][]Parameter)
	}

	for _, param := range op.Parameters {
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
