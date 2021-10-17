package gen

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/ast"
)

// schemaGen is used to convert openapi schemas into ast representation.
type schemaGen struct {
	// Spec is used to lookup for schema components
	// which is not in the 'refs' cache.
	spec *ogen.Spec

	// Contains schema nested objects.
	side []*ast.Schema

	// Global schema reference cache from *Generator (readonly).
	globalRefs map[string]*ast.Schema

	// Local schema reference cache.
	localRefs map[string]*ast.Schema
}

// Generate converts ogen.Schema into *ast.Schema.
//
// If ogen.Schema contains references to schema components,
// these referenced schemas will be saved in g.localRefs.
//
// If ogen.Schema contains nested objects, they will be
// collected in g.side slice.
func (g *schemaGen) Generate(name string, schema ogen.Schema) (*ast.Schema, error) {
	s, err := g.generate(pascal(name), schema, true, "")
	if err != nil {
		return nil, xerrors.Errorf("gen: %w", err)
	}
	return s, nil
}

func genericPostfix(name string) string {
	if idx := strings.Index(name, "."); idx > 0 {
		name = name[idx+1:]
	}
	return pascal(name)
}

func (g *schemaGen) generate(name string, schema ogen.Schema, root bool, ref string) (*ast.Schema, error) {
	if ref := schema.Ref; ref != "" {
		return g.ref(ref)
	}

	// sideEffect stores schema in g.localRefs or g.side if needed.
	sideEffect := func(s *ast.Schema) *ast.Schema {
		// Set validation fields.
		if schema.MultipleOf != nil {
			s.Validators.Int.MultipleOf = *schema.MultipleOf
			s.Validators.Int.MultipleOfSet = true
		}
		if schema.Maximum != nil {
			s.Validators.Int.Maximum = *schema.Maximum
			s.Validators.Int.MaximumSet = true
		}
		if schema.Minimum != nil {
			s.Validators.Int.Minimum = *schema.Minimum
			s.Validators.Int.MinimumSet = true
		}
		s.Validators.Int.ExclusiveMaximum = schema.ExclusiveMaximum
		s.Validators.Int.ExclusiveMinimum = schema.ExclusiveMinimum

		if schema.MaxItems != nil {
			s.Validators.Array.SetMaxLength(int(*schema.MaxItems))
		}
		if schema.MinItems != nil {
			s.Validators.Array.SetMinLength(int(*schema.MinItems))
		}

		if schema.MaxLength != nil {
			s.Validators.String.SetMaxLength(int(*schema.MaxLength))
		}
		if schema.MinLength != nil {
			s.Validators.String.SetMinLength(int(*schema.MinLength))
		}

		// s.Pattern = schema.Pattern

		// Referenced component, store it in g.localRefs.
		if ref != "" {
			// Reference pointed to a scalar type.
			// Wrap it with an alias using component name.
			if s.Is(ast.KindPrimitive, ast.KindArray, ast.KindPointer) {
				s = ast.Alias(name, s)
			}
			g.localRefs[ref] = s
			return s
		}

		// If schema it's a nested object (non-root)
		// and has a complex type (struct or alias) save it in g.side.
		if !root && !s.Is(ast.KindPrimitive, ast.KindArray, ast.KindPointer) {
			g.side = append(g.side, s)
		}
		return s
	}

	switch {
	case len(schema.Enum) > 0:
		typ, err := g.parseSimple(schema.Type, schema.Format)
		if err != nil {
			return nil, err
		}

		enum, err := ast.Enum(name, typ, schema.Enum)
		if err != nil {
			return nil, err
		}

		return sideEffect(enum), nil
	case len(schema.OneOf) > 0:
		return nil, &ErrNotImplemented{"oneOf"}
	case len(schema.AnyOf) > 0:
		return nil, &ErrNotImplemented{"anyOf"}
	case len(schema.AllOf) > 0:
		return nil, &ErrNotImplemented{"allOf"}
	}

	switch schema.Type {
	case "object":
		if len(schema.Properties) == 0 {
			return sideEffect(ast.Primitive("struct{}")), nil
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

		s := sideEffect(ast.Struct(name))
		s.Description = schema.Description
		if ref != "" {
			s.Doc = fmt.Sprintf("%s describes %s.", s.Name, ref)
		}
		for propName, propSchema := range schema.Properties {
			prop, err := g.generate(pascalMP(name, propName), propSchema, false, "")
			if err != nil {
				return nil, xerrors.Errorf("%s: %w", propName, err)
			}

			prop.Format = propSchema.Format

			canGeneric := prop.IsNumeric()
			switch prop.Primitive {
			case "bool", "string", "time.Time", "time.Duration", "uuid.UUID": // ok
				canGeneric = true
			}
			canGeneric = canGeneric && prop.Kind == ast.KindPrimitive
			if !required(propName) {
				if canGeneric {
					prop.Optional = true
					prop.GenericType = prop.GenericKind() + genericPostfix(prop.Primitive)
				} else {
					prop = ast.Pointer(prop)
				}
			}
			if propSchema.Nullable {
				if !canGeneric {
					return nil, xerrors.Errorf("not implemented: %w", &ErrNotImplemented{Name: "nullable"})
				}
				prop.Nil = true
				prop.GenericType = prop.GenericKind() + genericPostfix(prop.Primitive)
			}
			if prop.Generic() {
				switch prop.Format {
				case "uuid", "duration":
					prop.Format = ""
				}
			}

			s.Fields = append(s.Fields, ast.SchemaField{
				Name: pascalMP(propName),
				Type: prop,
				Tag:  propName,
			})
		}
		sort.SliceStable(s.Fields, func(i, j int) bool {
			return strings.Compare(s.Fields[i].Name, s.Fields[j].Name) < 0
		})
		return s, nil

	case "array":
		if schema.Items == nil {
			// Fallback to string.
			return sideEffect(ast.Array(ast.Primitive("string"))), nil
		}
		if len(schema.Properties) > 0 {
			return nil, xerrors.New("array cannot contain properties")
		}

		item, err := g.generate(name+"Item", *schema.Items, false, "")
		if err != nil {
			return nil, err
		}

		return sideEffect(ast.Array(item)), nil

	case "":
		return sideEffect(ast.Primitive("string")), nil

	default:
		simpleType, err := g.parseSimple(
			strings.ToLower(schema.Type),
			strings.ToLower(schema.Format),
		)
		if err != nil {
			return nil, xerrors.Errorf("parse: %w", err)
		}

		return sideEffect(ast.Primitive(simpleType)), nil
	}
}

func (g *schemaGen) ref(ref string) (*ast.Schema, error) {
	const prefix = "#/components/schemas/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, fmt.Errorf("invalid schema reference '%s'", ref)
	}

	if s, ok := g.globalRefs[ref]; ok {
		return s, nil
	}

	if s, ok := g.localRefs[ref]; ok {
		return s, nil
	}

	name := strings.TrimPrefix(ref, prefix)
	component, found := g.spec.Components.Schemas[name]
	if !found {
		return nil, xerrors.Errorf("component by reference '%s' not found", ref)
	}

	return g.generate(pascal(name), component, false, ref)
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
			"date-time": "time.Time",
			"time":      "time.Time",
			"date":      "time.Time",
			"duration":  "time.Duration",
			"password":  "string",
			"uuid":      "uuid.UUID",
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
