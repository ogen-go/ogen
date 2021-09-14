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

func (g *Generator) parseSchema(name string, schema ogen.Schema) (Schema, error) {
	s := Schema{
		Name:        name,
		Description: toFirstUpper(schema.Description),
		Implements:  map[string]struct{}{},
	}

	if s.Description != "" && !strings.HasSuffix(s.Description, ".") {
		s.Description += "."
	}

	if schema.Type != "object" {
		return Schema{}, fmt.Errorf("unexpected schema type: %s", schema.Type)
	}

	for pName, pSchema := range schema.Properties {
		pType, err := parseType(pSchema)
		if err != nil {
			return Schema{}, fmt.Errorf("property %s type: %w", pName, err)
		}

		s.Fields = append(s.Fields, SchemaField{
			Name: pascal(pName),
			Tag:  pName,
			Type: pType,
		})
	}

	sort.SliceStable(s.Fields, func(i, j int) bool {
		return strings.Compare(s.Fields[i].Name, s.Fields[j].Name) < 0
	})

	return s, nil
}

func (g *Generator) generateSchema(name string, schema ogen.Schema) (*Schema, error) {
	if schema.Ref != "" {
		return nil, fmt.Errorf("ref not supported")
	}

	switch schema.Type {
	case "object":
		s, err := g.parseSchema(name, schema)
		if err != nil {
			return nil, err
		}

		return &s, nil
	default:
		typ, err := parseType(schema)
		if err != nil {
			return nil, err
		}

		return &Schema{
			Name:        name,
			Description: schema.Description,
			Simple:      typ,
			Implements:  map[string]struct{}{},
		}, nil
	}
}
