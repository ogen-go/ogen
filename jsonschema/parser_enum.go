package jsonschema

import (
	"encoding/json"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
	"golang.org/x/exp/slices"

	"github.com/ogen-go/ogen/internal/xslices"
)

func inferJSONType(v json.RawMessage) (string, error) {
	d := jx.DecodeBytes(v)
	switch tt := d.Next(); tt {
	case jx.String:
		return "string", nil
	case jx.Number:
		return "number", nil
	case jx.Bool:
		return "bool", nil
	case jx.Null:
		return "", errors.Errorf("cannot infer type from %q", v)
	default:
		return "", errors.Errorf("invalid value %q", v)
	}
}

func parseEnumValues(s *Schema, rawValues []json.RawMessage) ([]any, error) {
	var values []any
	for _, raw := range rawValues {
		val, err := parseJSONValue(s, raw)
		if err != nil {
			return nil, errors.Wrapf(err, "parse value %q", raw)
		}
		values = append(values, val)
	}
	return values, nil
}

func parseJSONValue(root *Schema, v json.RawMessage) (any, error) {
	var parse func(s *Schema, d *jx.Decoder) (any, error)
	parse = func(s *Schema, d *jx.Decoder) (any, error) {
		switch next := d.Next(); next {
		case jx.Null:
			return nil, d.Null()
		case jx.String:
			return d.Str()
		case jx.Number:
			n, err := d.Num()
			if err != nil {
				return nil, err
			}
			if n.IsInt() {
				return n.Int64()
			}
			return n.Float64()
		case jx.Bool:
			return d.Bool()
		case jx.Array:
			var arr []any
			if err := d.Arr(func(d *jx.Decoder) error {
				var item *Schema
				if s != nil {
					item = s.Item
				}
				v, err := parse(item, d)
				if err != nil {
					return errors.Wrap(err, "array item")
				}
				arr = append(arr, v)
				return nil
			}); err != nil {
				return nil, err
			}
			return arr, nil
		case jx.Object:
			obj := map[string]any{}
			if err := d.ObjBytes(func(d *jx.Decoder, key []byte) error {
				v, err := parse(nil, d)
				if err != nil {
					return errors.Wrapf(err, "property %q", key)
				}
				obj[string(key)] = v
				return nil
			}); err != nil {
				return nil, err
			}
			return obj, nil
		default:
			return nil, errors.Errorf("unexpected type: %q", next)
		}
	}

	return parse(root, jx.DecodeBytes(v))
}

// See https://github.com/OAI/OpenAPI-Specification/blob/main/proposals/2019-10-31-Clarify-Nullable.md#if-a-schema-specifies-nullable-true-and-enum-1-2-3-does-that-schema-allow-null-values-see-1900.
func handleNullableEnum(s *Schema) {
	// Workaround: handle nullable enums correctly.
	//
	// Notice that nullable enum requires `null` in value list.
	//
	// Check that enum contains `null` value.
	s.Nullable = s.Nullable || slices.ContainsFunc(s.Enum, func(v any) bool {
		return v == nil
	})
	// Filter all `null`s.
	if s.Nullable {
		s.Enum = xslices.Filter(s.Enum, func(v any) bool {
			return v != nil
		})
	}
}
