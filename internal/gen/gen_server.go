package gen

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

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
				types := map[string]ParameterType{
					"query":  ParameterTypeQuery,
					"header": ParameterTypeHeader,
					"path":   ParameterTypePath,
					"cookie": ParameterCookie,
				}

				t, exists := types[strings.ToLower(param.In)]
				if !exists {
					return fmt.Errorf("unsupported parameter type %s", param.In)
				}

				if _, exists := method.Parameters[t]; !exists {
					method.Parameters[t] = []Parameter{}
				}

				if t == ParameterTypePath {
					exists, err := regexp.MatchString(fmt.Sprintf("{%s}", param.Name), p)
					if err != nil {
						return fmt.Errorf("match path param '%s': %w", param.Name, err)
					}

					if !exists {
						return fmt.Errorf("param '%s' not found in path '%s'", param.Name, p)
					}
				}

				paramType := param.Schema.Format
				if paramType == "" {
					paramType = param.Schema.Type
				}

				method.Parameters[t] = append(method.Parameters[t], Parameter{
					Name:       pascal(param.Name),
					SourceName: param.Name,
					Type:       paramType,
				})
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
