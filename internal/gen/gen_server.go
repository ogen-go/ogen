package gen

import (
	"fmt"
	"sort"
	"strings"
)

func (g *Generator) generateServer() error {
	for _, group := range g.spec.Paths {
		for _, pm := range group {
			method := serverMethodDef{
				Name: toFirstUpper(pm.OperationID),
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
		return strings.Compare(g.server.Methods[i].Name, g.server.Methods[j].Name) < 0
	})

	return nil
}
