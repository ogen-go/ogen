package gen

import (
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/internal/oas"
)

type schemaGen struct {
	side       []*ir.Type
	localRefs  map[string]*ir.Type
	globalRefs map[string]*ir.Type
}

func genericPostfix(name string) string {
	if idx := strings.Index(name, "."); idx > 0 {
		name = name[idx+1:]
	}
	return pascal(name)
}

func (g *schemaGen) generate(name string, schema *oas.Schema) (*ir.Type, error) {
	if ref := schema.Ref; ref != "" {
		if t, ok := g.globalRefs[ref]; ok {
			return t, nil
		}
		if t, ok := g.localRefs[ref]; ok {
			return t, nil
		}

		name = pascal(strings.TrimPrefix(ref, "#/components/schemas/"))
	}

	switch {
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

		if ref := t.Schema.Ref; ref != "" {
			if t.Is(ir.KindPrimitive, ir.KindArray) {
				t = ir.Alias(name, t)
			}

			g.localRefs[ref] = t
			return t
		}

		if t.Is(ir.KindStruct, ir.KindEnum, ir.KindSum) {
			g.side = append(g.side, t)
		}

		return t
	}

	switch schema.Type {
	case oas.Object:
		s := &ir.Type{
			Kind:   ir.KindStruct,
			Name:   name,
			Schema: schema,
		}

		s = side(s)

		for _, prop := range schema.Properties {
			typ, err := g.generate(pascalMP(name, prop.Name), prop.Schema)
			if err != nil {
				return nil, xerrors.Errorf("field '%s': %w", prop.Name, err)
			}
			v := ir.GenericVariant{
				Nullable: prop.Schema.Nullable,
				Optional: !prop.Required,
			}
			if v.Any() {
				if typ.CanGeneric() && !s.RecursiveTo(typ) {
					typ = ir.Generic(genericPostfix(typ.Go()),
						typ, v,
					)
					g.side = append(g.side, typ)
				} else if typ.IsArray() {
					// Using special case for array nil value if possible.
					switch {
					case v.OnlyOptional():
						typ.NilSemantic = ir.NilOptional
					case v.OnlyNullable():
						typ.NilSemantic = ir.NilNull
					default:
						typ = ir.Generic(genericPostfix(typ.Go()),
							typ, v,
						)
						g.side = append(g.side, typ)
					}
				} else {
					switch {
					case v.OnlyOptional():
						typ = typ.Pointer(ir.NilOptional)
					case v.OnlyNullable():
						typ = typ.Pointer(ir.NilNull)
					default:
						typ = ir.Generic(genericPostfix(typ.Go()),
							typ.Pointer(ir.NilNull), ir.GenericVariant{Optional: true},
						)
						g.side = append(g.side, typ)
					}
				}
			}
			if s.RecursiveTo(typ) {
				typ = typ.Pointer(ir.NilInvalid)
			}
			s.Fields = append(s.Fields, &ir.Field{
				Name: pascalMP(prop.Name),
				Type: typ,
				Tag: ir.Tag{
					JSON: prop.Name,
				},
				Spec: prop,
			})
		}

		return s, nil

	case oas.Array:
		array := &ir.Type{
			Kind:        ir.KindArray,
			Schema:      schema,
			NilSemantic: ir.NilInvalid,
		}

		ret := side(array)
		item, err := g.generate(name+"Item", schema.Item)
		if err != nil {
			return nil, err
		}

		array.Item = item
		return ret, nil

	case oas.String, oas.Integer, oas.Number, oas.Boolean:
		prim, err := g.primitive(name, schema)
		if err != nil {
			return nil, err
		}

		return side(prim), nil

	case oas.Empty:
		sum := &ir.Type{
			Name:   name,
			Kind:   ir.KindSum,
			Schema: schema,
		}
		for i, s := range schema.OneOf {
			t, err := g.generate(fmt.Sprintf("%s%d", name, i), s)
			if err != nil {
				return nil, xerrors.Errorf("oneOf[%d]: %w", i, err)
			}
			if !t.Is(ir.KindPrimitive) {
				return nil, xerrors.Errorf("%s: %w", name, &ErrNotImplemented{
					Name: "sum types for non-primitives",
				})
			}
			var result []rune
			for i, c := range t.Go() {
				if i == 0 {
					c = unicode.ToUpper(c)
				}
				result = append(result, c)
			}
			t.Name = string(result)
			sum.SumOf = append(sum.SumOf, t)
		}
		return side(sum), nil
	default:
		panic("unreachable")
	}
}

func (g *schemaGen) primitive(name string, schema *oas.Schema) (*ir.Type, error) {
	typ, err := parseSimple(schema)
	if err != nil {
		return nil, err
	}

	if len(schema.Enum) > 0 {
		if !typ.Is(ir.KindPrimitive) {
			return nil, xerrors.Errorf("unsupported enum type: '%s'", schema.Type)
		}

		var variants []*ir.EnumVariant
		for _, v := range schema.Enum {
			vstr := fmt.Sprintf("%v", v)
			if vstr == "" {
				vstr = "Empty"
			}

			variants = append(variants, &ir.EnumVariant{
				Name:  pascalMP(name, vstr),
				Value: v,
			})
		}

		return &ir.Type{
			Kind:         ir.KindEnum,
			Name:         name,
			Primitive:    typ.Primitive,
			EnumVariants: variants,
			Schema:       schema,
		}, nil
	}

	return typ, nil
}

func parseSimple(schema *oas.Schema) (*ir.Type, error) {
	typ, format := schema.Type, schema.Format
	switch typ {
	case oas.Integer:
		switch format {
		case oas.FormatInt32:
			return ir.Primitive(ir.Int32, schema), nil
		case oas.FormatInt64:
			return ir.Primitive(ir.Int64, schema), nil
		case oas.FormatNone:
			return ir.Primitive(ir.Int, schema), nil
		default:
			return nil, xerrors.Errorf("unexpected integer format: %q", format)
		}
	case oas.Number:
		switch format {
		case oas.FormatFloat:
			return ir.Primitive(ir.Float32, schema), nil
		case oas.FormatDouble, oas.FormatNone:
			return ir.Primitive(ir.Float64, schema), nil
		default:
			return nil, xerrors.Errorf("unexpected number format: %q", format)
		}
	case oas.String:
		switch format {
		case oas.FormatByte:
			return ir.Array(ir.Primitive(ir.Byte, nil), ir.NilInvalid, schema), nil
		case oas.FormatDateTime, oas.FormatDate, oas.FormatTime:
			return ir.Primitive(ir.Time, schema), nil
		case oas.FormatDuration:
			return ir.Primitive(ir.Duration, schema), nil
		case oas.FormatUUID:
			return ir.Primitive(ir.UUID, schema), nil
		case oas.FormatIP, oas.FormatIPv4, oas.FormatIPv6:
			return ir.Primitive(ir.IP, schema), nil
		case oas.FormatURI:
			return ir.Primitive(ir.URL, schema), nil
		case oas.FormatPassword, oas.FormatNone:
			return ir.Primitive(ir.String, schema), nil
		default:
			// return nil, xerrors.Errorf("unexpected string format: '%s'", format)
			return ir.Primitive(ir.String, schema), nil
		}
	case oas.Boolean:
		switch format {
		case oas.FormatNone:
			return ir.Primitive(ir.Bool, schema), nil
		default:
			return nil, xerrors.Errorf("unexpected bool format: %q", format)
		}
	default:
		return nil, xerrors.Errorf("unexpected type: %q", typ)
	}
}
