package gen

import (
	"sort"
	"strings"
)

func (g *Generator) generateServer() error {
	for _, group := range g.spec.Paths {
		for _, pm := range group {
			g.server.Methods = append(g.server.Methods, serverMethodDef{
				Name: toFirstUpper(pm.OperationID),
			})
		}
	}

	sort.SliceStable(g.server.Methods, func(i, j int) bool {
		return strings.Compare(g.server.Methods[i].Name, g.server.Methods[j].Name) < 0
	})

	return nil
}
