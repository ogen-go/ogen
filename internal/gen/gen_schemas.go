package gen

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/ogen-go/ogen"
)

type parseSimpleTypeParams struct {
	AllowArrays       bool
	AllowNestedArrays bool
}

func parseSimpleType(schema ogen.Schema, params parseSimpleTypeParams) (string, error) {
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

	switch t {
	case "array":
		if !params.AllowArrays {
			return "", fmt.Errorf("unsupported simple type %s", t)
		}

		if schema.Items == nil {
			return "", fmt.Errorf("items field is missed for array type")
		}

		itemType, err := parseSimpleType(*schema.Items, parseSimpleTypeParams{
			AllowArrays:       params.AllowNestedArrays,
			AllowNestedArrays: params.AllowNestedArrays,
		})
		if err != nil {
			return "", fmt.Errorf("array item type: %w", err)
		}

		return fmt.Sprintf("[]%s", itemType), nil
	default:
		formats, exists := simpleTypes[t]
		if !exists {
			return "", fmt.Errorf("unsupported simple type %s", t)
		}

		fType, exists := formats[f]
		if !exists {
			return "", fmt.Errorf("unsupported simple type %s format %s", t, f)
		}

		return fType, nil
	}
}

func parseType(schema ogen.Schema) (string, error) {
	t := strings.ToLower(schema.Type)
	f := strings.ToLower(schema.Format)

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

		itemType, err := parseType(*schema.Items)
		if err != nil {
			return "", fmt.Errorf("array item type: %w", err)
		}

		return fmt.Sprintf("[]%s", itemType), nil
	default:
		fType, err := parseSimpleType(schema, parseSimpleTypeParams{
			AllowArrays: false, // Arrays are already supported in the branch above.
		})
		if err != nil {
			return "", fmt.Errorf("unsupported type %s format %s", t, f)
		}

		return fType, nil
	}
}

func parseSchema(name string, schema ogen.Schema) (*schemaStructDef, error) {
	component := schemaStructDef{
		Name:        name,
		Description: toFirstUpper(schema.Description),
	}

	if !strings.HasSuffix(component.Description, ".") {
		component.Description += "."
	}

	for pName, pSchema := range schema.Properties {
		pType, err := parseType(pSchema)
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

func (g *Generator) generateSchemaComponents() error {
	for n, s := range g.spec.Components.Schemas {
		schema, err := parseSchema(n, s)
		if err != nil {
			return fmt.Errorf("parse component %s: %w", n, err)
		}

		schema.Path = path.Join("#/components/schemas", n)
		g.schemas = append(g.schemas, *schema)
	}

	sort.SliceStable(g.schemas, func(i, j int) bool {
		return strings.Compare(g.schemas[i].Name, g.schemas[j].Name) < 0
	})

	return nil
}

func (g *Generator) schemaComponentByRef(ref string) string {
	for _, c := range g.schemas {
		if c.Path == ref {
			return c.Name
		}
	}

	return ""
}
