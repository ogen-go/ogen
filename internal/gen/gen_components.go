package gen

import (
	"path"
	"sort"
	"strings"
)

func (g *Generator) generateComponents() error {
	for n, s := range g.spec.Components.Schemas {
		component := componentStructDef{
			Name:        n,
			Description: toFirstUpper(s.Description),
			Path:        path.Join("#/components/schemas", n),
		}

		if !strings.HasSuffix(component.Description, ".") {
			component.Description += "."
		}

		for pName, pSchema := range s.Properties {
			f := field{
				Name:    pascal(pName),
				TagName: pName,
				Type:    pSchema.Type,
			}

			switch f.Type {
			case "integer":
				f.Type = pSchema.Format
			}

			component.Fields = append(component.Fields, f)
		}

		sort.SliceStable(component.Fields, func(i, j int) bool {
			return strings.Compare(component.Fields[i].Name, component.Fields[j].Name) < 0
		})

		g.components = append(g.components, component)
	}

	sort.SliceStable(g.components, func(i, j int) bool {
		return strings.Compare(g.components[i].Name, g.components[j].Name) < 0
	})

	return nil
}
