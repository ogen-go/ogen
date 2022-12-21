package ogenreflect

import (
	"reflect"
	"strconv"
)

// RuntimeOperation stores the operation information.
type RuntimeOperation struct {
	// Name is the ogen operation name. It is guaranteed to be unique and not empty.
	Name string
	// ID is the spec operation ID, if any.
	ID string
	// Types stores the type information for the operation.
	Types OperationTypes
}

// OperationTypes holds the operation types.
type OperationTypes struct {
	// Request is the operation request type.
	Request RequestType
	// Params stores the operation parameters types.
	Params ParametersType
	// Responses stores the operation responses types.
	Responses ResponsesType
}

// IsRequest checks if the type is the operation request type.
func (t OperationTypes) IsRequest(v any) bool {
	r := t.Request
	if r.Type == nil {
		// Operation has no request.
		return false
	}
	if len(r.Implementations) == 0 {
		return reflect.TypeOf(v) == r.Type
	}
	for _, impl := range r.Implementations {
		if reflect.TypeOf(v) == impl {
			return true
		}
	}
	return false
}

// IsParam checks if the type is the operation param type.
func (t OperationTypes) IsParam(v any) bool {
	for _, impl := range t.Params.Map {
		if reflect.TypeOf(v) == impl.Type {
			return true
		}
	}
	return false
}

// IsResponse checks if the type is the operation response type.
func (t OperationTypes) IsResponse(v any) bool {
	r := t.Responses
	if len(r.Implementations) == 0 {
		return reflect.TypeOf(v) == r.Type
	}
	for _, impl := range r.Implementations {
		if reflect.TypeOf(v) == impl {
			return true
		}
	}
	return false
}

// RequestType holds the request type information.
type RequestType struct {
	// Type is the request type.
	//
	// Type is nil if the operation has no request body.
	//
	// If the requestBody defines multiple content types, Type is the interface type, implemented
	// by all Implementations types.
	Type reflect.Type

	// Implementations is the request type implementations.
	Implementations []reflect.Type

	// Contents stores the request contents by pattern.
	Contents Contents
}

// ParameterType holds the parameter type information.
type ParameterType struct {
	// Type is the parameter type.
	Type reflect.Type
	// Name is the spec parameter name.
	Name string
	// In is the parameter location.
	In string
	// Style is the parameter style.
	Style string
	// Explode is true if the parameter is exploded.
	Explode bool
	// Required is true if the parameter is required.
	Required bool
}

// ParametersType holds the parameters type information.
type ParametersType struct {
	// StructType is the parameters struct type.
	//
	// StructType is nil if the operation has no parameters.
	StructType reflect.Type
	// Map stores the operation parameters types.
	Map ParameterMap[ParameterType]
}

// ResponsesType holds the response type information.
type ResponsesType struct {
	// Type is the response type.
	//
	// If operation defines multiple content types, Type is the interface type, implemented
	// by all Implementations types.
	Type reflect.Type

	// Implementations is the request type implementations.
	Implementations []reflect.Type

	// PatternMap stores the response contents by pattern.
	//
	// If element is empty, the response has no content for the pattern.
	PatternMap map[string]ResponseType
}

// FindResponse returns the matching contents for the given status code.
func (r ResponsesType) FindResponse(code int) (rt ResponseType, ok bool) {
	if code < 100 || code > 599 {
		return rt, false
	}
	rt, ok = r.PatternMap[strconv.Itoa(code)]
	if ok {
		return rt, true
	}
	switch code / 100 {
	case 1:
		rt, ok = r.PatternMap["1XX"]
	case 2:
		rt, ok = r.PatternMap["2XX"]
	case 3:
		rt, ok = r.PatternMap["3XX"]
	case 4:
		rt, ok = r.PatternMap["4XX"]
	case 5:
		rt, ok = r.PatternMap["5XX"]
	}
	if ok {
		return rt, true
	}
	rt, ok = r.PatternMap["default"]
	return rt, ok
}

// ResponseType is the response type.
type ResponseType struct {
	// Headers stores the response headers.
	Headers map[string]ParameterType
	// Contents stores the response contents.
	Contents Contents
}

// Contents is the request or response contents.
//
// The key is the content type pattern.
type Contents map[string]reflect.Type
