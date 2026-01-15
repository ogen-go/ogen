package gen

import (
	"fmt"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/location"
)

func (g *schemaGen) primitive(name string, schema *jsonschema.Schema) (*ir.Type, error) {
	t := g.parseSimple(schema)

	if len(schema.Enum) > 0 {
		return g.enum(name, t, schema)
	}

	return t, nil
}

func (g *schemaGen) enum(name string, t *ir.Type, schema *jsonschema.Schema) (*ir.Type, error) {
	if !t.Is(ir.KindPrimitive) {
		return nil, errors.Wrapf(&ErrNotImplemented{"complex enum type"}, "type %s", t.String())
	}

	// We accept 2 types of enums: ints and strings. However, for formatted
	// string enums, we don't want to allow time/date/date-time formats as they
	// require special handling
	if f := schema.Format; f != "" {
		if !t.IsNumeric() && t.Schema.Type != jsonschema.String {
			return nil, errors.Wrapf(&ErrNotImplemented{"enum format"}, "format %q", f)
		}

		// Reject time-related formats for string enums until we properly handle them
		if t.Schema.Type == jsonschema.String {
			switch f {
			case "date", "time", "date-time", "http-date":
				return nil, errors.Wrapf(&ErrNotImplemented{"enum format"}, "format %q", f)
			}
		}
	}

	if err := g.validateEnumValues(schema); err != nil {
		return nil, errors.Wrap(err, "validate enum")
	}

	nameGen, err := enumVariantNameGen(name, schema.Enum)
	if err != nil {
		return nil, errors.Wrap(err, "choose strategy")
	}

	var variants []*ir.EnumVariant
	for idx, v := range schema.Enum {
		variantName, err := nameGen(v, idx)
		if err != nil {
			return nil, errors.Wrapf(err, "variant %q [%d]", fmt.Sprintf("%v", v), idx)
		}

		variants = append(variants, &ir.EnumVariant{
			Name:  variantName,
			Value: v,
		})
	}

	return &ir.Type{
		Kind:         ir.KindEnum,
		Name:         name,
		Primitive:    t.Primitive,
		EnumVariants: variants,
		Schema:       schema,
	}, nil
}

func (g *schemaGen) validateEnumValues(s *jsonschema.Schema) error {
	reportErr := func(idx int, err error) error {
		pos, ok := s.Pointer.Field("enum").Index(idx).Position()
		if !ok {
			return err
		}
		return &location.Error{
			File: s.File(),
			Pos:  pos,
			Err:  err,
		}
	}

	switch typ := s.Type; typ {
	case jsonschema.Object, jsonschema.Array, jsonschema.Empty:
		return &ErrNotImplemented{Name: "non-primitive enum"}
	case jsonschema.Integer:
		for idx, val := range s.Enum {
			if _, ok := val.(int64); !ok {
				return reportErr(idx, errors.Errorf("enum value should be an integer, got %T", val))
			}
		}
		return nil
	case jsonschema.Number:
		for idx, val := range s.Enum {
			switch val.(type) {
			case int64, float64:
			default:
				return reportErr(idx, errors.Errorf("enum value should be a number, got %T", val))
			}
		}
		return nil
	case jsonschema.String:
		for idx, val := range s.Enum {
			if _, ok := val.(string); !ok {
				return reportErr(idx, errors.Errorf("enum value should be a string, got %T", val))
			}
		}
		return nil
	case jsonschema.Boolean:
		for idx, val := range s.Enum {
			if _, ok := val.(bool); !ok {
				return reportErr(idx, errors.Errorf("enum value should be a boolean, got %T", val))
			}
		}
		return nil
	case jsonschema.Null:
		for idx, val := range s.Enum {
			if val != nil {
				return reportErr(idx, errors.Errorf("enum value should be a null, got %T", val))
			}
		}
		return nil
	default:
		panic(fmt.Sprintf("unexpected schema type %q", typ))
	}
}

func (g *schemaGen) parseSimple(schema *jsonschema.Schema) *ir.Type {
	mapping := TypeFormatMapping()

	// TODO(tdakkota): check ContentEncoding field
	t, found := mapping[schema.Type][schema.Format]
	if !found {
		// Fallback to default.
		t = mapping[schema.Type][""]
	}

	return ir.Primitive(t, schema)
}

func TypeFormatMapping() map[jsonschema.SchemaType]map[string]ir.PrimitiveType {
	return map[jsonschema.SchemaType]map[string]ir.PrimitiveType{
		jsonschema.Integer: {
			"": ir.Int,

			// FIXME(tdakkota): add decoder for int8, int16, uint8, uint16 to jx.
			"int8":   ir.Int8,
			"int16":  ir.Int16,
			"int32":  ir.Int32,
			"int64":  ir.Int64,
			"uint":   ir.Uint,
			"uint8":  ir.Uint8,
			"uint16": ir.Uint16,
			"uint32": ir.Uint32,
			"uint64": ir.Uint64,

			// See https://github.com/ogen-go/ogen/issues/307.
			"unix":         ir.Time,
			"unix-seconds": ir.Time,
			"unix-nano":    ir.Time,
			"unix-micro":   ir.Time,
			"unix-milli":   ir.Time,
		},
		jsonschema.Number: {
			"float":   ir.Float32,
			"double":  ir.Float64,
			"int32":   ir.Int32,
			"int64":   ir.Int64,
			"decimal": ir.Decimal,
			"":        ir.Float64,
		},
		jsonschema.String: {
			"byte":      ir.ByteSlice,
			"base64":    ir.ByteSlice,
			"date-time": ir.Time,
			"date":      ir.Time,
			"time":      ir.Time,
			"http-date": ir.Time,
			"duration":  ir.Duration,
			"uuid":      ir.UUID,
			"mac":       ir.MAC,
			"ip":        ir.IP,
			"ipv4":      ir.IP,
			"ipv6":      ir.IP,
			"uri":       ir.URL,
			"password":  ir.String,
			"email":     ir.String,
			"binary":    ir.String,
			"hostname":  ir.String,
			"":          ir.String,

			// Custom format, see https://github.com/ogen-go/ogen/issues/309.
			"int":   ir.Int,
			"int8":  ir.Int8,
			"int16": ir.Int16,
			"int32": ir.Int32,
			"int64": ir.Int64,

			"uint":   ir.Uint,
			"uint8":  ir.Uint8,
			"uint16": ir.Uint16,
			"uint32": ir.Uint32,
			"uint64": ir.Uint64,
			// See https://github.com/ogen-go/ogen/issues/307.
			"unix":         ir.Time,
			"unix-seconds": ir.Time,
			"unix-nano":    ir.Time,
			"unix-micro":   ir.Time,
			"unix-milli":   ir.Time,
			// See https://github.com/ogen-go/ogen/issues/957.
			"float32": ir.Float32,
			"float64": ir.Float64,
			"decimal": ir.Decimal,
		},
		jsonschema.Boolean: {
			"": ir.Bool,
		},
		jsonschema.Null: {
			"": ir.Null,
		},
	}
}
