package gen

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/ast"
)

func (g *Generator) generateSchema(name string, schema ogen.Schema) (*ast.Schema, error) {
	if ref := schema.Ref; ref != "" {
		componentName, err := componentName(ref)
		if err != nil {
			return nil, xerrors.Errorf("invalid schema reference: %s", ref)
		}

		s, found := g.schemas[pascal(componentName)]
		if !found {
			refSchema, found := g.spec.Components.Schemas[componentName]
			if !found {
				return nil, xerrors.Errorf("component by reference '%s' not found", ref)
			}

			s, err := g.generateSchema(pascal(componentName), refSchema)
			if err != nil {
				return nil, err
			}

			return s, nil
		}

		return s, nil
	}

	switch {
	case len(schema.OneOf) > 0:
		return ast.Primitive("interface{}"), nil
	case len(schema.AnyOf) > 0:
		return ast.Primitive("interface{}"), nil
	case len(schema.AllOf) > 0:
		return ast.Primitive("interface{}"), nil
	}

	switch schema.Type {
	case "object":
		if len(schema.Properties) == 0 {
			return ast.Primitive("struct{}"), nil
		}

		if schema.Items != nil {
			return nil, xerrors.New("object cannot contain 'items' field")
		}

		required := func(name string) bool {
			for _, p := range schema.Required {
				if p == name {
					return true
				}
			}
			return false
		}

		s := ast.Struct(name)
		s.Description = schema.Description
		g.schemas[s.Name] = s
		for propName, propSchema := range schema.Properties {
			prop, err := g.generateSchema(name+pascalMP(propName), propSchema)
			if err != nil {
				return nil, xerrors.Errorf("%s: %w", propName, err)
			}

			typ := prop.Type()
			if !required(propName) {
				typ = "*" + typ
			}

			s.Fields = append(s.Fields, ast.SchemaField{
				Name: pascalMP(propName),
				Tag:  propName,
				Type: typ,
			})
		}
		sort.SliceStable(s.Fields, func(i, j int) bool {
			return strings.Compare(s.Fields[i].Name, s.Fields[j].Name) < 0
		})
		return s, nil

	case "array":
		if schema.Items == nil {
			if g.opt.debugSkipUnspecified {
				return nil, xerrors.Errorf("skipping unspecified items: %w", errSkipSchema)
			}
			return nil, xerrors.New("items must be specified")
		}
		if len(schema.Properties) > 0 {
			return nil, xerrors.New("array cannot contain properties")
		}

		item, err := g.generateSchema(name+"Item", *schema.Items)
		if err != nil {
			return nil, err
		}

		return ast.Array(item), nil

	case "":
		if g.opt.debugSkipUnspecified {
			return nil, xerrors.Errorf("skipping unspecified type: %w", errSkipSchema)
		}
		return nil, xerrors.New("type must be specified")

	default:
		simpleType, err := g.parseSimple(
			strings.ToLower(schema.Type),
			strings.ToLower(schema.Format),
		)
		if err != nil {
			return nil, xerrors.Errorf("parse: %w", err)
		}

		return ast.Primitive(simpleType), nil
	}
}

func (g *Generator) parseSimple(typ, format string) (string, error) {
	simpleTypes := map[string]map[string]string{
		"integer": {
			"int32": "int32",
			"int64": "int64",
			"":      "int",
		},
		"number": {
			"float":  "float32",
			"double": "float64",
			"":       "float",
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
		return "", fmt.Errorf("unsupported type: '%s'", typ)
	}

	fType, exists := formats[format]
	if !exists {
		// Fallback to string.
		if typ == "string" {
			return "string", nil
		}

		return "", xerrors.Errorf("unsupported format '%s' for type '%s'", format, typ)
	}

	return fType, nil
}
