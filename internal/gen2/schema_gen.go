package gen

import (
	ast "github.com/ogen-go/ogen/internal/ast2"
	"github.com/ogen-go/ogen/internal/ir"
	"golang.org/x/xerrors"
)

type schemaGen struct {
	side       []*ir.Type
	localRefs  map[string]*ir.Type
	globalRefs map[string]*ir.Type
}

func (g *schemaGen) generate(name string, schema *ast.Schema) (*ir.Type, error) {
	if ref := schema.Ref; ref != "" {
		if t, ok := g.globalRefs[ref]; ok {
			return t, nil
		}
		if t, ok := g.localRefs[ref]; ok {
			return t, nil
		}
	}

	switch {
	case len(schema.OneOf) > 0:
		return nil, &ErrNotImplemented{"oneOf"}
	case len(schema.AnyOf) > 0:
		return nil, &ErrNotImplemented{"anyOf"}
	case len(schema.AllOf) > 0:
		return nil, &ErrNotImplemented{"allOf"}
	}

	side := func(t *ir.Type) *ir.Type {
		// Set validation fields.
		if schema.MultipleOf != nil {
			t.Validators.Int.MultipleOf = *schema.MultipleOf
			t.Validators.Int.MultipleOfSet = true
		}
		if schema.Maximum != nil {
			t.Validators.Int.Max = *schema.Maximum
			t.Validators.Int.MaxSet = true
		}
		if schema.Minimum != nil {
			t.Validators.Int.Min = *schema.Minimum
			t.Validators.Int.MinSet = true
		}
		t.Validators.Int.MaxExclusive = schema.ExclusiveMaximum
		t.Validators.Int.MinExclusive = schema.ExclusiveMinimum

		if schema.MaxItems != nil {
			t.Validators.Array.SetMaxLength(int(*schema.MaxItems))
		}
		if schema.MinItems != nil {
			t.Validators.Array.SetMinLength(int(*schema.MinItems))
		}

		if schema.MaxLength != nil {
			t.Validators.String.SetMaxLength(int(*schema.MaxLength))
		}
		if schema.MinLength != nil {
			t.Validators.String.SetMinLength(int(*schema.MinLength))
		}

		if ref := t.Spec.Ref; ref != "" {
			if t.Is(ir.KindPrimitive, ir.KindArray) {
				t = ir.Alias(name, t)
			}

			g.localRefs[ref] = t
			return t
		}

		if t.Is(ir.KindStruct, ir.KindEnum) {
			g.side = append(g.side, t)
		}

		return t
	}

	switch schema.Type {
	case ast.Object:
		s := &ir.Type{
			Kind: ir.KindStruct,
			Name: name,
			Spec: schema,
		}

		s = side(s)

		for _, prop := range schema.Properties {
			typ, err := g.generate(pascalMP(name, prop.Name), prop.Schema)
			if err != nil {
				return nil, xerrors.Errorf("field '%s': %w", prop.Name, err)
			}

			if prop.Optional {
				typ = &ir.Type{
					Kind:      ir.KindPointer,
					PointerTo: typ,
				}
			}

			f := &ir.StructField{
				Name: pascalMP(prop.Name),
				Type: typ,
				Tag:  prop.Name,
			}

			s.Fields = append(s.Fields, f)
		}

		return s, nil

	case ast.Array:
		array := &ir.Type{
			Kind: ir.KindArray,
			Spec: schema,
		}

		ret := side(array)
		item, err := g.generate(name+"Item", schema.Item)
		if err != nil {
			return nil, err
		}

		array.Item = item
		return ret, nil

	case ast.String, ast.Integer, ast.Number, ast.Boolean:
		prim, err := g.primitive(name, schema)
		if err != nil {
			return nil, err
		}

		return side(prim), nil

	default:
		panic("unreachable")
	}
}

func (g *schemaGen) primitive(name string, schema *ast.Schema) (*ir.Type, error) {
	typ, err := parseSimple(schema)
	if err != nil {
		return nil, err
	}

	if len(schema.Enum) > 0 {
		if !typ.Is(ir.KindPrimitive) {
			return nil, xerrors.Errorf("unsupported enum type: '%s'", schema.Type)
		}

		return &ir.Type{
			Kind:       ir.KindEnum,
			Name:       name,
			Primitive:  typ.Primitive,
			EnumValues: schema.Enum,
			Spec:       schema,
		}, nil
	}

	return typ, nil
}

func parseSimple(schema *ast.Schema) (*ir.Type, error) {
	typ, format := schema.Type, schema.Format
	switch typ {
	case ast.Integer:
		switch format {
		case "int32":
			return ir.Primitive(ir.Int32, schema), nil
		case "int64":
			return ir.Primitive(ir.Int64, schema), nil
		case "":
			return ir.Primitive(ir.Int, schema), nil
		default:
			return nil, xerrors.Errorf("unexpected integer format: '%s'", format)
		}
	case ast.Number:
		switch format {
		case "float":
			return ir.Primitive(ir.Float32, schema), nil
		case "double", "":
			return ir.Primitive(ir.Float64, schema), nil
		default:
			return nil, xerrors.Errorf("unexpected number format: '%s'", format)
		}
	case ast.String:
		switch format {
		case "byte":
			return ir.Array(ir.Primitive(ir.Byte, nil), schema), nil
		case "date-time", "time", "date":
			return ir.Primitive(ir.Time, schema), nil
		case "duration":
			return ir.Primitive(ir.Duration, schema), nil
		case "uuid":
			return ir.Primitive(ir.UUID, schema), nil
		case "ipv4", "ipv6", "ip":
			return ir.Primitive(ir.IP, schema), nil
		case "uri":
			return ir.Primitive(ir.URL, schema), nil
		case "password", "":
			return ir.Primitive(ir.String, schema), nil
		default:
			// return nil, xerrors.Errorf("unexpected string format: '%s'", format)
			return ir.Primitive(ir.String, schema), nil
		}
	case ast.Boolean:
		switch format {
		case "":
			return ir.Primitive(ir.Bool, schema), nil
		default:
			return nil, xerrors.Errorf("unexpected bool format: '%s'", format)
		}
	default:
		return nil, xerrors.Errorf("unexpected type: '%s'", typ)
	}
}
