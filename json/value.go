package json

import (
	json "github.com/json-iterator/go"
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
