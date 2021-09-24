package gen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ogen-go/ogen"
)

func (g *Generator) generateSchema(name string, schema ogen.Schema) (*Schema, error) {
	return g.generateSchemaWithOpts(name, schema, true)
}

func (g *Generator) generateSchemaWithOpts(name string, schema ogen.Schema, allowNestedArrays bool) (*Schema, error) {
	if ref := schema.Ref; ref != "" {
		componentName, err := componentName(ref)
		if err != nil {
			return nil, fmt.Errorf("invalid schema reference: %s", ref)
		}

		s, found := g.schemas[componentName]
		if !found {
			return nil, fmt.Errorf("component by reference '%s' not found", ref)
		}

		return s, nil
	}

	switch schema.Type {
	case "object":
		if len(schema.Properties) == 0 {
			return nil, fmt.Errorf("object must contain at least one property")
		}

		required := func(name string) bool {
			for _, p := range schema.Required {
				if p == name {
					return true
				}
			}
			return false
		}

		s := g.createSchemaStruct(name)
		s.Description = schema.Description
		g.schemas[s.Name] = s
		for propName, propSchema := range schema.Properties {
			if !allowNestedArrays && propSchema.Type == "array" {
				return nil, fmt.Errorf("properties: %s: nested array not allowed", propName)
			}

			if !required(propName) {
				return nil, fmt.Errorf("properties: %s: optional properties not supported", propName)
			}

			prop, err := g.generateSchema(name+pascal(propName), propSchema)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", propName, err)
			}

			s.Fields = append(s.Fields, SchemaField{
				Name: pascal(propName),
				Tag:  propName,
				Type: prop.Type(),
			})
		}
		sort.SliceStable(s.Fields, func(i, j int) bool {
			return strings.Compare(s.Fields[i].Name, s.Fields[j].Name) < 0
		})
		return s, nil
	case "array":
		if schema.Items == nil {
			return nil, fmt.Errorf("empty items field")
		}

		if !allowNestedArrays && schema.Items.Type == "array" {
			return nil, fmt.Errorf("item: nested arrays not allowed")
		}

		item, err := g.generateSchema(name+"Item", *schema.Items)
		if err != nil {
			return nil, err
		}

		return g.createSchemaArray(name, item), nil
	default:
		simpleType, err := parseSimple(
			strings.ToLower(schema.Type),
			strings.ToLower(schema.Format),
		)
		if err != nil {
			return nil, err
		}

		return g.createSchemaPrimitive(name, simpleType), nil
	}
}

func parseSimple(typ, format string) (string, error) {
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

	formats, exists := simpleTypes[typ]
	if !exists {
		return "", fmt.Errorf("unsupported simple type %s", typ)
	}

	fType, exists := formats[format]
	if !exists {
		return "", fmt.Errorf("unsupported simple type %s format %s", typ, format)
	}

	return fType, nil
}
