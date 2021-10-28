package json

import (
	"github.com/ogen-go/json"
)

const (
	// InvalidValue invalid JSON element
	InvalidValue = json.InvalidValue
	// StringValue JSON element "string"
	StringValue = json.StringValue
	// NumberValue JSON element 100 or 0.10
	NumberValue = json.NumberValue
	// NilValue JSON element null
	NilValue = json.NilValue
	// BoolValue JSON element true or false
	BoolValue = json.BoolValue
	// ArrayValue JSON element []
	ArrayValue = json.ArrayValue
	// ObjectValue JSON element {}
	ObjectValue = json.ObjectValue
)

// TypeStr returns string representation of s.
func TypeStr(v json.ValueType) string {
	switch v {
	case InvalidValue:
		return "invalid"
	case NumberValue:
		return "number"
	case NilValue:
		return "nil"
	case ArrayValue:
		return "array"
	case ObjectValue:
		return "object"
	case BoolValue:
		return "bool"
	case StringValue:
		return "string"
	default:
		return "unknown"
	}
}
