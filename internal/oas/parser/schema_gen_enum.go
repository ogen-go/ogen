package parser

import (
	"encoding/json"

	jsoniter "github.com/ogen-go/json"
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
		iter = jsoniter.ParseBytes(jsoniter.ConfigDefault, v)
		next = iter.WhatIsNext()
	)

	if next == jsoniter.NilValue {
		return nil, errNullValue
	}

	str := func(t jsoniter.ValueType) string {
		switch t {
		case jsoniter.InvalidValue:
			return "invalid"
		case jsoniter.StringValue:
			return "string"
		case jsoniter.NumberValue:
			return "number"
		case jsoniter.NilValue:
			return "null"
		case jsoniter.BoolValue:
			return "bool"
		case jsoniter.ArrayValue:
			return "array"
		case jsoniter.ObjectValue:
			return "object"
		default:
			return "unknown"
		}
	}

	switch typ {
	case oas.String:
		if next != jsoniter.StringValue {
			expect, actual := typ, str(next)
			return nil, xerrors.Errorf("expect type '%s', got '%s'", expect, actual)
		}
		return iter.ReadString(), nil
	case oas.Integer:
		if next != jsoniter.NumberValue {
			expect, actual := typ, str(next)
			return nil, xerrors.Errorf("expect type '%s', got '%s'", expect, actual)
		}

		if _, err := iter.ReadNumber().Float64(); err == nil {
			expect, actual := typ, "float"
			return nil, xerrors.Errorf("expect type '%s', got '%s'", expect, actual)
		}

		return iter.ReadNumber().Int64()
	case oas.Number:
		if next != jsoniter.NumberValue {
			expect, actual := typ, str(next)
			return nil, xerrors.Errorf("expect type '%s', got '%s'", expect, actual)
		}

		if _, err := iter.ReadNumber().Int64(); err == nil {
			expect, actual := typ, "int"
			return nil, xerrors.Errorf("expect type '%s', got '%s'", expect, actual)
		}

		return iter.ReadNumber().Float64()
	case oas.Boolean:
		if next != jsoniter.BoolValue {
			expect, actual := typ, str(next)
			return nil, xerrors.Errorf("expect type '%s', got '%s'", expect, actual)
		}

		return iter.ReadBool(), nil
	default:
		return nil, xerrors.Errorf("unexpected type: '%s'", typ)
	}
}
