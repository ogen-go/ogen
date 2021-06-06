package gen

import (
	"fmt"
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

			g.server.Methods = append(g.server.Methods, method)
		}
	}

	sort.SliceStable(g.server.Methods, func(i, j int) bool {
		return strings.Compare(g.server.Methods[i].Path, g.server.Methods[j].Path) < 0 ||
			strings.Compare(g.server.Methods[i].HTTPMethod, g.server.Methods[j].HTTPMethod) < 0
	})

	return nil
}
