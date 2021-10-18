package ast

import (
	"encoding/json"

	jsoniter "github.com/json-iterator/go"
	"golang.org/x/xerrors"
)

var errNullValue = xerrors.New("json null value")

func parseJSONValue(typ string, v json.RawMessage) (interface{}, error) {
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
	case "string":
		if next != jsoniter.StringValue {
			expect, actual := typ, str(next)
			return nil, xerrors.Errorf("expect type '%s', got '%s'", expect, actual)
		}
		return iter.ReadString(), nil
	case "int", "int8", "int16", "int32", "int64":
		if next != jsoniter.NumberValue {
			expect, actual := typ, str(next)
			return nil, xerrors.Errorf("expect type '%s', got '%s'", expect, actual)
		}

		if _, err := iter.ReadNumber().Float64(); err == nil {
			expect, actual := typ, "float"
			return nil, xerrors.Errorf("expect type '%s', got '%s'", expect, actual)
		}

		return iter.ReadNumber().Int64()
	case "float32", "float64":
		if next != jsoniter.NumberValue {
			expect, actual := typ, str(next)
			return nil, xerrors.Errorf("expect type '%s', got '%s'", expect, actual)
		}

		if _, err := iter.ReadNumber().Int64(); err == nil {
			expect, actual := typ, "int"
			return nil, xerrors.Errorf("expect type '%s', got '%s'", expect, actual)
		}

		return iter.ReadNumber().Float64()
	case "bool":
		if next != jsoniter.BoolValue {
			expect, actual := typ, str(next)
			return nil, xerrors.Errorf("expect type '%s', got '%s'", expect, actual)
		}

		return iter.ReadBool(), nil
	default:
		return nil, xerrors.Errorf("unexpected type: '%s'", typ)
	}
}
