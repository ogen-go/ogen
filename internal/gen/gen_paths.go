package gen

import (
	"fmt"
	"sort"
	"strings"
)

func (g *Generator) generatePaths() error {
	for p, group := range g.spec.Paths {
		pathGroup := pathGroupDef{
			Path: p,
		}

		for m, pm := range group {
			pathMethod := pathMethodDef{
				HTTPMethod: strings.ToUpper(m),
			}

			serverMethod := g.serverMethod(pm.OperationID)
			if serverMethod == "" {
				return fmt.Errorf("server method not found for %s", pm.OperationID)
			}

			pathMethod.ServerMethod = serverMethod

			for _, content := range pm.RequestBody.Content {
				name := g.componentByRef(content.Schema.Ref)
				if name == "" {
					return fmt.Errorf("ref %s not found", content.Schema.Ref)
				}

				pathMethod.RequestType = name
			}

			pathGroup.Methods = append(pathGroup.Methods, pathMethod)
		}

		sort.SliceStable(pathGroup.Methods, func(i, j int) bool {
			return strings.Compare(pathGroup.Methods[i].HTTPMethod, pathGroup.Methods[j].HTTPMethod) < 0
		})

		g.groups = append(g.groups, pathGroup)
	}

	sort.SliceStable(g.groups, func(i, j int) bool {
		return strings.Compare(g.groups[i].Path, g.groups[j].Path) < 0
	})

	return nil
}

func (g *Generator) serverMethod(operationID string) string {
	for _, m := range g.server.Methods {
		if m.OperationID == operationID {
			return m.Name
		}
	}

	return ""
}
