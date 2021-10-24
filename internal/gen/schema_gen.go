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
		s.Format = schema.Format

		// Set validation fields.
		if schema.MultipleOf != nil {
			s.Validators.Int.MultipleOf = *schema.MultipleOf
			s.Validators.Int.MultipleOfSet = true
		}
		if schema.Maximum != nil {
			s.Validators.Int.Max = *schema.Maximum
			s.Validators.Int.MaxSet = true
		}
		if schema.Minimum != nil {
			s.Validators.Int.Min = *schema.Minimum
			s.Validators.Int.MinSet = true
		}
		s.Validators.Int.MaxExclusive = schema.ExclusiveMaximum
		s.Validators.Int.MinExclusive = schema.ExclusiveMinimum

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
			if s.Is(ast.KindPrimitive, ast.KindArray, ast.KindPointer, ast.KindGeneric) {
				s = ast.Alias(name, s)
			}
			g.localRefs[ref] = s
			return s
		}

		// If schema it's a nested object (non-root)
		// and has a complex type (struct or alias) save it in g.side.
		if !root && !s.Is(ast.KindPrimitive, ast.KindArray, ast.KindPointer, ast.KindGeneric) {
			g.side = append(g.side, s)
		}
		return s
	}

	switch {
	case len(schema.Enum) > 0:
		simple, err := g.parseSimple(schema.Type, schema.Format)
		if err != nil {
			return nil, err
		}

		if simple.Kind != ast.KindPrimitive {
			return nil, xerrors.Errorf("unsupported enum type '%s' format '%s'", schema.Type, schema.Format)
		}

		enum, err := ast.Enum(name, simple.Primitive, schema.Enum)
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
			return sideEffect(ast.Primitive(ast.EmptyStruct)), nil
		}

		if schema.Items != nil {
			return nil, xerrors.New("object cannot contain 'items' field")
		}
		optional := func(name string) bool {
			for _, p := range schema.Required {
				if p == name {
					return false
				}
			}
			return true
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
			v := ast.GenericVariant{
				Nullable: propSchema.Nullable,
				Optional: optional(propName),
			}
			if v.Any() {
				if prop.CanGeneric() && !s.RecursiveTo(prop) {
					// Box value with generic wrapper.
					prop.Format = propSchema.Format
					prop = ast.Generic(
						genericPostfix(prop.Type()),
						prop,
						v,
					)
					g.side = append(g.side, prop)
				} else if prop.IsArray() {
					// Using special case for array nil value if possible.
					switch {
					case v.OnlyOptional():
						prop.NilSemantic = ast.NilOptional
					case v.OnlyNullable():
						prop.NilSemantic = ast.NilNull
					default:
						// TODO(ernado): fallback to boxing
						return nil, xerrors.Errorf("%s: %w", ref, &ErrNotImplemented{Name: "optional nullable array"})
					}
				} else {
					switch {
					case v.OnlyOptional():
						prop = ast.Pointer(prop, ast.NilOptional)
					case v.OnlyNullable():
						prop = ast.Pointer(prop, ast.NilNull)
					default:
						panic("unreachable")
					}
				}
			}
			if s.RecursiveTo(prop) {
				prop = ast.Pointer(prop, ast.NilInvalid)
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
			return sideEffect(ast.Array(ast.Primitive(ast.String))), nil
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
		return sideEffect(ast.Primitive(ast.String)), nil

	default:
		simple, err := g.parseSimple(
			strings.ToLower(schema.Type),
			strings.ToLower(schema.Format),
		)
		if err != nil {
			return nil, xerrors.Errorf("parse: %w", err)
		}

		return sideEffect(simple), nil
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

func (g *schemaGen) parseSimple(typ, format string) (*ast.Schema, error) {
	switch typ {
	case "integer":
		switch format {
		case "int32":
			return ast.Primitive(ast.Int32), nil
		case "int64":
			return ast.Primitive(ast.Int64), nil
		case "":
			return ast.Primitive(ast.Int), nil
		default:
			return nil, xerrors.Errorf("unexpected integer format: '%s'", format)
		}
	case "number":
		switch format {
		case "float":
			return ast.Primitive(ast.Float32), nil
		case "double", "":
			return ast.Primitive(ast.Float64), nil
		default:
			return nil, xerrors.Errorf("unexpected number format: '%s'", format)
		}
	case "string":
		switch format {
		case "byte":
			return ast.Array(ast.Primitive(ast.Byte)), nil
		case "date-time", "time", "date":
			return ast.Primitive(ast.Time), nil
		case "duration":
			return ast.Primitive(ast.Duration), nil
		case "uuid":
			return ast.Primitive(ast.UUID), nil
		case "ipv4", "ipv6", "ip":
			return ast.Primitive(ast.IP), nil
		case "uri":
			return ast.Primitive(ast.URL), nil
		case "password", "":
			return ast.Primitive(ast.String), nil
		default:
			// return nil, xerrors.Errorf("unexpected string format: '%s'", format)
			return ast.Primitive(ast.String), nil
		}
	case "boolean":
		switch format {
		case "":
			return ast.Primitive(ast.Bool), nil
		default:
			return nil, xerrors.Errorf("unexpected bool format: '%s'", format)
		}
	default:
		return nil, xerrors.Errorf("unexpected type: '%s'", typ)
	}
}
