package parser

import (
	"encoding/json"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"

	"github.com/ogen-go/ogen/internal/oas"
)

func parseEnumValues(typ oas.SchemaType, rawValues []json.RawMessage) ([]interface{}, error) {
	var (
		values []interface{}
		unique = map[interface{}]struct{}{}
	)
	for _, raw := range rawValues {
		val, err := parseJSONValue(typ, raw)
		if err != nil {
			return nil, errors.Wrapf(err, "parse value %q", raw)
		}

		if _, found := unique[val]; found {
			return nil, errors.Errorf("duplicate enum value: '%v'", val)
		}

		unique[val] = struct{}{}
		values = append(values, val)
	}
	return values, nil
}

func parseJSONValue(typ oas.SchemaType, v json.RawMessage) (interface{}, error) {
	var (
		d    = jx.DecodeBytes(v)
		next = d.Next()
	)
	if next == jx.Null {
		return nil, nil
	}
	switch typ {
	case oas.String:
		if next != jx.String {
			return nil, errors.Errorf("expected type %q, got %q", typ, next)
		}
		return d.Str()
	case oas.Integer:
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
	case oas.Number:
		if next != jx.Number {
			return nil, errors.Errorf("expected type %q, got %q", typ, next)
		}
		n, err := d.Num()
		if err != nil {
			return nil, err
		}
		return n.Float64()
	case oas.Boolean:
		if next != jx.Bool {
			return nil, errors.Errorf("expected type %q, got %q", typ, next)
		}
		return d.Bool()
	default:
		return nil, errors.Errorf("unexpected type: %q", typ)
	}
}
