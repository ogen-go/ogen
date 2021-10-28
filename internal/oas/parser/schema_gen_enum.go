package parser

import (
	"encoding/json"

	"github.com/ogen-go/jir"
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
		iter = jir.ParseBytes(jir.Default, v)
		next = iter.WhatIsNext()
	)
	if next == jir.Nil {
		return nil, errNullValue
	}
	switch typ {
	case oas.String:
		if next != jir.String {
			return nil, xerrors.Errorf("expected type %q, got %q", typ, next)
		}
		return iter.Str(), nil
	case oas.Integer:
		if next != jir.Number {
			return nil, xerrors.Errorf("expected type %q, got %q", typ, next)
		}
		if _, err := iter.ReadNumber().Float64(); err == nil {
			return nil, xerrors.Errorf("expected type %q, got %q", typ, next)
		}
		return iter.ReadNumber().Int64()
	case oas.Number:
		if next != jir.Number {
			return nil, xerrors.Errorf("expected type %q, got %q", typ, next)
		}
		if _, err := iter.ReadNumber().Int64(); err == nil {
			return nil, xerrors.Errorf("expected type %q, got %q", typ, next)
		}
		return iter.ReadNumber().Float64()
	case oas.Boolean:
		if next != jir.Bool {
			return nil, xerrors.Errorf("expected type %q, got %q", typ, next)
		}
		return iter.Bool(), nil
	default:
		return nil, xerrors.Errorf("unexpected type: '%s'", typ)
	}
}
