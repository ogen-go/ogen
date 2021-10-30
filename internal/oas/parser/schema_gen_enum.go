package parser

import (
	"encoding/json"

	"github.com/ogen-go/jx"
	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen/internal/oas"
)

func parseEnumValues(typ oas.SchemaType, rawValues []json.RawMessage) ([]interface{}, error) {
	var (
		values []interface{}
		uniq   = map[interface{}]struct{}{}
	)
	for _, raw := range rawValues {
		val, err := parseJSONValue(typ, raw)
		if err != nil {
			if xerrors.Is(err, errNullValue) {
				continue
			}
			return nil, xerrors.Errorf("parse value '%s': %w", raw, err)
		}

		if _, found := uniq[val]; found {
			return nil, xerrors.Errorf("duplicate enum value: '%v'", val)
		}

		uniq[val] = struct{}{}
		values = append(values, val)
	}

	return values, nil
}

var errNullValue = xerrors.New("json null value")

func parseJSONValue(typ oas.SchemaType, v json.RawMessage) (interface{}, error) {
	var (
		d    = jx.DecodeBytes(v)
		next = d.Next()
	)
	if next == jx.Nil {
		return nil, errNullValue
	}
	switch typ {
	case oas.String:
		if next != jx.String {
			return nil, xerrors.Errorf("expected type %q, got %q", typ, next)
		}
		return d.Str()
	case oas.Integer:
		if next != jx.Number {
			return nil, xerrors.Errorf("expected type %q, got %q", typ, next)
		}
		n, err := d.Number()
		if err != nil {
			return nil, err
		}
		if _, err := n.Float64(); err == nil {
			return nil, xerrors.Errorf("expected type %q, got %q", typ, next)
		}
		return n.Int64()
	case oas.Number:
		if next != jx.Number {
			return nil, xerrors.Errorf("expected type %q, got %q", typ, next)
		}
		n, err := d.Number()
		if err != nil {
			return nil, err
		}
		if _, err := n.Int64(); err == nil {
			return nil, xerrors.Errorf("expected type %q, got %q", typ, next)
		}
		return n.Float64()
	case oas.Boolean:
		if next != jx.Bool {
			return nil, xerrors.Errorf("expected type %q, got %q", typ, next)
		}
		return d.Bool()
	default:
		return nil, xerrors.Errorf("unexpected type: '%s'", typ)
	}
}
