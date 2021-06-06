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
			serverMethod := g.serverMethod(pm.OperationID)
			if serverMethod == "" {
				return fmt.Errorf("server method not found for %s", pm.OperationID)
			}

			pathGroup.Methods = append(pathGroup.Methods, pathMethodDef{
				HTTPMethod:   strings.ToUpper(m),
				ServerMethod: serverMethod,
			})
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
