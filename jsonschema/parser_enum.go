package jsonschema

import (
	"encoding/json"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"

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
		next := d.Next()
		if next == jx.Null {
			// We do not check nullability here because enum with null value is completely valid
			// even if it is not nullable.
			return nil, nil
		}
		switch typ := s.Type; typ {
		case String:
			if next != jx.String {
				return nil, errors.Errorf("expected type %q, got %q", typ, next)
			}
			return d.Str()
		case Integer:
			if next != jx.Number {
				return nil, errors.Errorf("expected type %q, got %q", typ, next)
			}
			n, err := d.Num()
			if err != nil {
				return nil, err
			}
			if !n.IsInt() {
				return nil, errors.New("expected integer, got float")
			}
			return n.Int64()
		case Number:
			if next != jx.Number {
				return nil, errors.Errorf("expected type %q, got %q", typ, next)
			}
			n, err := d.Num()
			if err != nil {
				return nil, err
			}
			return n.Float64()
		case Boolean:
			if next != jx.Bool {
				return nil, errors.Errorf("expected type %q, got %q", typ, next)
			}
			return d.Bool()
		case Array:
			if next != jx.Array {
				return nil, errors.Errorf("expected type %q, got %q", typ, next)
			}
			if s.Item == nil {
				return nil, errors.New("can't validate untyped array item")
			}
			var arr []any
			if err := d.Arr(func(d *jx.Decoder) error {
				v, err := parse(s.Item, d)
				if err != nil {
					return errors.Wrap(err, "validate item")
				}
				arr = append(arr, v)
				return nil
			}); err != nil {
				return nil, err
			}
			return arr, nil
		default:
			return nil, errors.Errorf("unexpected type: %q", typ)
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
	s.Nullable = s.Nullable || xslices.ContainsFunc(s.Enum, func(v any) bool {
		return v == nil
	})
	// Filter all `null`s.
	if s.Nullable {
		xslices.Filter(&s.Enum, func(v any) bool {
			return v != nil
		})
	}
}
