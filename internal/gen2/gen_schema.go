package gen

import (
	ast "github.com/ogen-go/ogen/internal/ast2"
	"github.com/ogen-go/ogen/internal/ir"
	"golang.org/x/xerrors"
)

func (g *Generator) generateSchema(name string, schema *ast.Schema) (*ir.Type, error) {
	if schema.Ref != "" {
		name = pascal(schema.Ref)
	}

	switch schema.Type {
	case ast.Object:
		return g.object2IR(name, schema)
	case ast.Array:
		return g.array2IR(name, schema)
	case ast.String, ast.Number, ast.Integer, ast.Boolean:
		return g.primitive2IR(name, schema)
	default:
		panic("unreachable")
	}
}

func (g *Generator) object2IR(name string, schema *ast.Schema) (*ir.Type, error) {
	s := &ir.Type{
		Kind: ir.KindStruct,
		Name: name,
		Spec: schema,
	}

	for _, field := range schema.Fields {
		typ, err := g.generateSchema(pascalMP(name, field.Name), field.Schema)
		if err != nil {
			return nil, xerrors.Errorf("field '%s': %w", field.Name, err)
		}

		if field.Schema.Type == ast.Object {
			if field.Schema.Ref != "" {
				g.structs[field.Schema.Ref] = typ
			} else {
				g.refs[typ.Name] = typ
			}
		}

		if field.Optional {
			typ = &ir.Type{
				Kind:      ir.KindPointer,
				PointerTo: typ,
			}
		}

		f := &ir.StructField{
			Name: pascalMP(field.Name),
			Type: typ,
			Tag:  field.Name,
		}

		s.Fields = append(s.Fields, f)
	}

	return s, nil
}

func (g *Generator) array2IR(name string, schema *ast.Schema) (*ir.Type, error) {
	item, err := g.generateSchema(name+"Item", schema.Item)
	if err != nil {
		return nil, err
	}

	return &ir.Type{
		Kind: ir.KindArray,
		Item: item,
		Spec: schema,
	}, nil
}

func (g *Generator) primitive2IR(name string, schema *ast.Schema) (*ir.Type, error) {
	typ, err := parseSimple(schema)
	if err != nil {
		return nil, err
	}

	if len(schema.EnumValues) > 0 {
		if !typ.Is(ir.KindPrimitive) {
			return nil, xerrors.Errorf("unsupported enum type: '%s'", schema.Type)
		}

		return &ir.Type{
			Kind:       ir.KindEnum,
			Primitive:  typ.Primitive,
			EnumValues: schema.EnumValues,
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
