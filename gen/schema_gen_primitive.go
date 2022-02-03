package gen

import (
	"fmt"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/internal/oas"
)

func (g *schemaGen) primitive(name string, schema *oas.Schema) (*ir.Type, error) {
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

func parseSimple(schema *oas.Schema) (*ir.Type, error) {
	mapping := map[oas.SchemaType]map[oas.Format]ir.PrimitiveType{
		oas.Integer: {
			oas.FormatInt32: ir.Int32,
			oas.FormatInt64: ir.Int64,
			oas.FormatNone:  ir.Int,
		},
		oas.Number: {
			oas.FormatFloat:  ir.Float32,
			oas.FormatDouble: ir.Float64,
			oas.FormatNone:   ir.Float64,
			oas.FormatInt32:  ir.Int32,
			oas.FormatInt64:  ir.Int64,
		},
		oas.String: {
			oas.FormatByte:     ir.ByteSlice,
			oas.FormatDateTime: ir.Time,
			oas.FormatDate:     ir.Time,
			oas.FormatTime:     ir.Time,
			oas.FormatDuration: ir.Duration,
			oas.FormatUUID:     ir.UUID,
			oas.FormatIP:       ir.IP,
			oas.FormatIPv4:     ir.IP,
			oas.FormatIPv6:     ir.IP,
			oas.FormatURI:      ir.URL,
			oas.FormatPassword: ir.String,
			oas.FormatNone:     ir.String,
		},
		oas.Boolean: {
			oas.FormatNone: ir.Bool,
		},
	}

	t, found := mapping[schema.Type][schema.Format]
	if !found {
		// Return string type for unknown string formats.
		if schema.Type == oas.String {
			return ir.Primitive(ir.String, schema), nil
		}
		return nil, errors.Errorf("unexpected %q format: %q", schema.Type, schema.Format)
	}

	return ir.Primitive(t, schema), nil
}
