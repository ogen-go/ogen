package gen

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/ernado/ogen"
)

func getType(schema ogen.ComponentSchema) (string, error) {
	t := strings.ToLower(schema.Type)
	f := strings.ToLower(schema.Format)

	simpleTypes := map[string]map[string]string{
		"integer": {
			"int32": "int32",
			"int64": "int64",
		},
		"number": {
			"float":  "float32",
			"double": "float64",
		},
		"string": {
			"":          "string",
			"byte":      "[]byte",
			"date":      "time.Time",
			"date-time": "time.Time",
			"password":  "string",
			// TODO: support binary format
		},
		"boolean": {
			"": "bool",
		},
	}

	if formats, exists := simpleTypes[t]; exists {
		if fType, formatExists := formats[f]; formatExists {
			return fType, nil
		}

		return "", fmt.Errorf("unsupported type %s format %s", t, f)
	}

	switch t {
	case "object":
		if schema.Ref == "" {
			return "", fmt.Errorf("nested object fields supported only by ref")
		}

		return path.Base(schema.Ref), nil
	case "array":
		if schema.Items == nil {
			return "", fmt.Errorf("items field is missed for array type")
		}

		itemType, err := getType(*schema.Items)
		if err != nil {
			return "", fmt.Errorf("array item type: %w", err)
		}

		return fmt.Sprintf("[]%s", itemType), nil
	default:
		if formats, exists := simpleTypes[t]; exists {
			if fType, formatExists := formats[f]; !formatExists {
				return fType, nil
			}

			return "", fmt.Errorf("unsupported type %s format %s", t, f)
		}

		return "", fmt.Errorf("unsupported type %s", t)
	}
}

func parseComponent(name string, schema ogen.ComponentSchema) (*componentStructDef, error) {
	component := componentStructDef{
		Name:        name,
		Description: toFirstUpper(schema.Description),
		Path:        path.Join("#/components/schemas", name),
	}

	if !strings.HasSuffix(component.Description, ".") {
		component.Description += "."
	}

	for pName, pSchema := range schema.Properties {
		pType, err := getType(pSchema)
		if err != nil {
			return nil, fmt.Errorf("property %s type: %w", pName, err)
		}

		f := field{
			Name:    pascal(pName),
			TagName: pName,
			Type:    pType,
		}

		component.Fields = append(component.Fields, f)
	}

	sort.SliceStable(component.Fields, func(i, j int) bool {
		return strings.Compare(component.Fields[i].Name, component.Fields[j].Name) < 0
	})

	return &component, nil
}

func (g *Generator) generateComponents() error {
	for n, s := range g.spec.Components.Schemas {
		component, err := parseComponent(n, s)
		if err != nil {
			return fmt.Errorf("parse component %s: %w", n, err)
		}

		g.components = append(g.components, *component)
	}

	sort.SliceStable(g.components, func(i, j int) bool {
		return strings.Compare(g.components[i].Name, g.components[j].Name) < 0
	})

	return nil
}

func (g *Generator) componentByRef(ref string) string {
	for _, c := range g.components {
		if c.Path == ref {
			return c.Name
		}
	}

	return ""
}
