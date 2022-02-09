package gen

import (
	"fmt"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/jsonschema"
)

func (g *schemaGen) primitive(name string, schema *jsonschema.Schema) (*ir.Type, error) {
	t := parseSimple(schema)

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
				variantName = name + "_" + pascalSpecial(vstr)
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

func parseSimple(schema *jsonschema.Schema) *ir.Type {
	mapping := map[jsonschema.SchemaType]map[string]ir.PrimitiveType{
		jsonschema.Integer: {
			"int32": ir.Int32,
			"int64": ir.Int64,
			"":      ir.Int,
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
			"date-time": ir.Time,
			"date":      ir.Time,
			"time":      ir.Time,
			"duration":  ir.Duration,
			"uuid":      ir.UUID,
			"ip":        ir.IP,
			"ipv4":      ir.IP,
			"ipv6":      ir.IP,
			"uri":       ir.URL,
			"password":  ir.String,
			"":          ir.String,
		},
		jsonschema.Boolean: {
			"": ir.Bool,
		},
	}

	t, found := mapping[schema.Type][schema.Format]
	if !found {
		// Fallback to default.
		t = mapping[schema.Type][""]
	}

	return ir.Primitive(t, schema)
}
