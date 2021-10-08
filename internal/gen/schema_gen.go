package gen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/ast"
	"golang.org/x/xerrors"
)

type schemaGen struct {
	spec *ogen.Spec
	side []*ast.Schema
	refs map[string]*ast.Schema
}

func (g *schemaGen) Generate(name string, schema ogen.Schema) (*ast.Schema, error) {
	return g.generate(pascal(name), schema, true, "")
}

func (g *schemaGen) GenerateRef(ref string) (*ast.Schema, error) {
	return g.ref(ref)
}

func (g *schemaGen) generate(name string, schema ogen.Schema, root bool, ref string) (s *ast.Schema, err error) {
	if ref := schema.Ref; ref != "" {
		return g.ref(ref)
	}

	switch {
	case len(schema.Enum) > 0:
		return nil, ErrEnumsNotImplemented
	case len(schema.OneOf) > 0:
		return nil, ErrOneOfNotImplemented
	case len(schema.AnyOf) > 0:
		return nil, ErrAnyOfNotImplemented
	case len(schema.AllOf) > 0:
		return nil, ErrAllOfNotImplemented
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
		if ref != "" {
			g.refs[ref] = s
		} else if !root {
			g.side = append(g.side, s)
		}

		for propName, propSchema := range schema.Properties {
			prop, err := g.generate(pascalMP(name, propName), propSchema, false, "")
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
			// Fallback to string.
			return ast.Array(ast.Primitive("string")), nil
		}
		if len(schema.Properties) > 0 {
			return nil, xerrors.New("array cannot contain properties")
		}

		item, err := g.generate(name+"Item", *schema.Items, false, "")
		if err != nil {
			return nil, err
		}

		return ast.Array(item), nil

	case "":
		return ast.Primitive("string"), nil

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

func (g *schemaGen) ref(ref string) (*ast.Schema, error) {
	if !strings.HasPrefix(ref, "#/components/schemas/") {
		return nil, fmt.Errorf("invalid schema reference '%s'", ref)
	}

	if s, ok := g.refs[ref]; ok {
		return s, nil
	}

	specComponentName, err := componentName(ref)
	if err != nil {
		return nil, xerrors.Errorf("invalid schema reference: %s", ref)
	}

	component, found := g.spec.Components.Schemas[specComponentName]
	if !found {
		return nil, xerrors.Errorf("component by reference '%s' not found", ref)
	}

	return g.generate(pascal(specComponentName), component, false, ref)
}

func (g *schemaGen) parseSimple(typ, format string) (string, error) {
	simpleTypes := map[string]map[string]string{
		"integer": {
			"int32": "int32",
			"int64": "int64",
			"":      "int",
		},
		"number": {
			"float":  "float32",
			"double": "float64",
			"":       "float64",
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
