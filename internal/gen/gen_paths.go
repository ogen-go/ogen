package gen

import (
	"sort"
	"strings"
)

func (g *Generator) generatePaths() error {
	for p, group := range g.spec.Paths {
		pathGroup := pathGroupDef{
			Path: p,
		}

		for m := range group {
			pathGroup.Methods = append(pathGroup.Methods, pathMethodDef{
				Method: strings.ToUpper(m),
			})
		}

		sort.SliceStable(pathGroup.Methods, func(i, j int) bool {
			return strings.Compare(pathGroup.Methods[i].Method, pathGroup.Methods[j].Method) < 0
		})

		g.groups = append(g.groups, pathGroup)
	}

	sort.SliceStable(g.groups, func(i, j int) bool {
		return strings.Compare(g.groups[i].Path, g.groups[j].Path) < 0
	})

	return nil
}
