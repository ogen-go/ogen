package gen

import (
	"fmt"
	"strconv"
	"unicode/utf8"

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
	if f := schema.Format; f != "" && !t.IsNumeric() {
		return nil, errors.Wrapf(&ErrNotImplemented{"enum format"}, "format %q", f)
	}
	if err := g.validateEnumValues(schema); err != nil {
		return nil, errors.Wrap(err, "validate enum")
	}

	type namingStrategy int
	const (
		pascalName namingStrategy = iota
		pascalSpecialName
		cleanSuffix
		indexSuffix
		_lastStrategy
	)

	vstrCache := make(map[int]string, len(schema.Enum))
	nameEnum := func(s namingStrategy, idx int, v any) (string, error) {
		vstr, ok := vstrCache[idx]
		if !ok {
			vstr = fmt.Sprintf("%v", v)
			if vstr == "" {
				vstr = "Empty"
			}
			vstrCache[idx] = vstr
		}

		switch s {
		case pascalName:
			return pascal(name, vstr)
		case pascalSpecialName:
			return pascalSpecial(name, vstr)
		case cleanSuffix:
			return name + "_" + cleanSpecial(vstr), nil
		case indexSuffix:
			return name + "_" + strconv.Itoa(idx), nil
		default:
			panic(unreachable(s))
		}
	}

	isException := func(start namingStrategy) bool {
		if start == pascalName {
			// This code is called when vstrCache is fully populated, so it's ok.
			for _, v := range vstrCache {
				if v == "" {
					continue
				}

				// Do not use pascal strategy for enum values starting with special characters.
				//
				// This rule is created to be able to distinguish
				// between negative and positive numbers in this case:
				//
				// enum:
				//   - '1'
				//   - '-2'
				//   - '3'
				//   - '-4'
				firstRune, _ := utf8.DecodeRuneInString(v)
				if firstRune == utf8.RuneError {
					panic(fmt.Sprintf("invalid enum value: %q", v))
				}

				_, isFirstCharSpecial := namedChar[firstRune]
				if isFirstCharSpecial {
					return true
				}
			}
		}

		return false
	}

	chosenStrategy, err := func() (namingStrategy, error) {
	nextStrategy:
		for strategy := pascalName; strategy < _lastStrategy; strategy++ {
			// Treat enum type name as duplicate to prevent collisions.
			names := map[string]struct{}{
				name: {},
			}
			for idx, v := range schema.Enum {
				k, err := nameEnum(strategy, idx, v)
				if err != nil {
					continue nextStrategy
				}
				if _, ok := names[k]; ok {
					continue nextStrategy
				}
				names[k] = struct{}{}
			}
			if isException(strategy) {
				continue nextStrategy
			}
			return strategy, nil
		}
		return 0, errors.Errorf("unable to generate variant names for enum %q", name)
	}()
	if err != nil {
		return nil, errors.Wrap(err, "choose strategy")
	}

	var variants []*ir.EnumVariant
	for idx, v := range schema.Enum {
		variantName, err := nameEnum(chosenStrategy, idx, v)
		if err != nil {
			return nil, errors.Wrapf(err, "variant %q [%d]", vstrCache[idx], idx)
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
			File: s.Pointer.File(),
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
			"float":  ir.Float32,
			"double": ir.Float64,
			"int32":  ir.Int32,
			"int64":  ir.Int64,
			"":       ir.Float64,
		},
		jsonschema.String: {
			"byte":      ir.ByteSlice,
			"base64":    ir.ByteSlice,
			"date-time": ir.Time,
			"date":      ir.Time,
			"time":      ir.Time,
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
		},
		jsonschema.Boolean: {
			"": ir.Bool,
		},
		jsonschema.Null: {
			"": ir.Null,
		},
	}
}
