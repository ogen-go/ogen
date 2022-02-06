package gen

import (
	"fmt"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/jsonschema"
)

func (g *schemaGen) primitive(name string, schema *jsonschema.Schema) (*ir.Type, error) {
	t, err := parseSimple(schema)
	if err != nil {
		return nil, err
	}

	if len(schema.Enum) > 0 {
		if !t.Is(ir.KindPrimitive) {
			return nil, errors.Errorf("unsupported enum type: %q", schema.Type)
		}

		hasDuplicateNames := func() bool {
			names := map[string]struct{}{}
			for _, v := range schema.Enum {
				vstr := fmt.Sprintf("%v", v)
				if vstr == "" {
					vstr = "Empty"
				}

				k := pascalSpecial(name, vstr)
				if _, ok := names[k]; ok {
					return true
				}
				names[k] = struct{}{}
			}

			return false
		}()

		var variants []*ir.EnumVariant
		for _, v := range schema.Enum {
			vstr := fmt.Sprintf("%v", v)
			if vstr == "" {
				vstr = "Empty"
			}

			var variantName string
			if hasDuplicateNames {
				variantName = name + "_" + vstr
			} else {
				variantName = pascalSpecial(name, vstr)
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

	return t, nil
}

func parseSimple(schema *jsonschema.Schema) (*ir.Type, error) {
	mapping := map[jsonschema.SchemaType]map[jsonschema.Format]ir.PrimitiveType{
		jsonschema.Integer: {
			jsonschema.FormatInt32: ir.Int32,
			jsonschema.FormatInt64: ir.Int64,
			jsonschema.FormatNone:  ir.Int,
		},
		jsonschema.Number: {
			jsonschema.FormatFloat:  ir.Float32,
			jsonschema.FormatDouble: ir.Float64,
			jsonschema.FormatNone:   ir.Float64,
			jsonschema.FormatInt32:  ir.Int32,
			jsonschema.FormatInt64:  ir.Int64,
		},
		jsonschema.String: {
			jsonschema.FormatByte:     ir.ByteSlice,
			jsonschema.FormatDateTime: ir.Time,
			jsonschema.FormatDate:     ir.Time,
			jsonschema.FormatTime:     ir.Time,
			jsonschema.FormatDuration: ir.Duration,
			jsonschema.FormatUUID:     ir.UUID,
			jsonschema.FormatIP:       ir.IP,
			jsonschema.FormatIPv4:     ir.IP,
			jsonschema.FormatIPv6:     ir.IP,
			jsonschema.FormatURI:      ir.URL,
			jsonschema.FormatPassword: ir.String,
			jsonschema.FormatNone:     ir.String,
		},
		jsonschema.Boolean: {
			jsonschema.FormatNone: ir.Bool,
		},
	}

	t, found := mapping[schema.Type][schema.Format]
	if !found {
		// Return string type for unknown string formats.
		if schema.Type == jsonschema.String {
			return ir.Primitive(ir.String, schema), nil
		}
		return nil, errors.Errorf("unexpected %q format: %q", schema.Type, schema.Format)
	}

	return ir.Primitive(t, schema), nil
}
