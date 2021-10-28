package parser

import (
	"encoding/json"

	j "github.com/ogen-go/json"
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
		iter = j.ParseBytes(j.ConfigDefault, v)
		next = iter.WhatIsNext()
	)

	if next == j.NilValue {
		return nil, errNullValue
	}

	str := func(t j.ValueType) string {
		switch t {
		case j.InvalidValue:
			return "invalid"
		case j.StringValue:
			return "string"
		case j.NumberValue:
			return "number"
		case j.NilValue:
			return "null"
		case j.BoolValue:
			return "bool"
		case j.ArrayValue:
			return "array"
		case j.ObjectValue:
			return "object"
		default:
			return "unknown"
		}
	}

	switch typ {
	case oas.String:
		if next != j.StringValue {
			expect, actual := typ, str(next)
			return nil, xerrors.Errorf("expect type '%s', got '%s'", expect, actual)
		}
		return iter.ReadString(), nil
	case oas.Integer:
		if next != j.NumberValue {
			expect, actual := typ, str(next)
			return nil, xerrors.Errorf("expect type '%s', got '%s'", expect, actual)
		}

		if _, err := iter.ReadNumber().Float64(); err == nil {
			expect, actual := typ, "float"
			return nil, xerrors.Errorf("expect type '%s', got '%s'", expect, actual)
		}

		return iter.ReadNumber().Int64()
	case oas.Number:
		if next != j.NumberValue {
			expect, actual := typ, str(next)
			return nil, xerrors.Errorf("expect type '%s', got '%s'", expect, actual)
		}

		if _, err := iter.ReadNumber().Int64(); err == nil {
			expect, actual := typ, "int"
			return nil, xerrors.Errorf("expect type '%s', got '%s'", expect, actual)
		}

		return iter.ReadNumber().Float64()
	case oas.Boolean:
		if next != j.BoolValue {
			expect, actual := typ, str(next)
			return nil, xerrors.Errorf("expect type '%s', got '%s'", expect, actual)
		}

		return iter.ReadBool(), nil
	default:
		return nil, xerrors.Errorf("unexpected type: '%s'", typ)
	}
}
