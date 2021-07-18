package gen

import (
	"fmt"
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
	for p, group := range g.spec.Paths {
		for m, pm := range group {
			method := serverMethodDef{
				Name:        toFirstUpper(pm.OperationID),
				OperationID: pm.OperationID,
				Path:        p,
				HTTPMethod:  strings.ToUpper(m),
			}

			for _, content := range pm.RequestBody.Content {
				name := g.componentByRef(content.Schema.Ref)
				if name == "" {
					return fmt.Errorf("ref %s not found", content.Schema.Ref)
				}

				method.RequestType = name
			}

			for status, resp := range pm.Responses {
				if status != "200" {
					continue
				}

				for _, content := range resp.Content {
					name := g.componentByRef(content.Schema.Ref)
					if name == "" {
						return fmt.Errorf("ref %s not found", content.Schema.Ref)
					}

					method.ResponseType = name
				}
			}

			if len(pm.Parameters) != 0 {
				method.Parameters = make(map[ParameterType][]Parameter)
			}

			for _, param := range pm.Parameters {
				parameter, err := parseParameter(param, p)
				if err != nil {
					return fmt.Errorf("parse method %s parameter %s: %w", pm.OperationID, param.Name, err)
				}

				if _, exists := method.Parameters[parameter.In]; !exists {
					method.Parameters[parameter.In] = []Parameter{}
				}

				method.Parameters[parameter.In] = append(method.Parameters[parameter.In], *parameter)
			}

			g.server.Methods = append(g.server.Methods, method)
		}
	}

	sort.SliceStable(g.server.Methods, func(i, j int) bool {
		return strings.Compare(g.server.Methods[i].Path, g.server.Methods[j].Path) < 0 ||
			strings.Compare(g.server.Methods[i].HTTPMethod, g.server.Methods[j].HTTPMethod) < 0
	})

	return nil
}
