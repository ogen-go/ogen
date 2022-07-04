package openapi

import (
	"github.com/go-json-experiment/json"

	"github.com/ogen-go/ogen/jsonschema"
)

// MediaType is Media Type Object.
type MediaType struct {
	Schema   *jsonschema.Schema
	Example  json.RawValue
	Examples map[string]*Example
	Encoding map[string]*Encoding
}

// Encoding is Encoding Type Object.
type Encoding struct {
	ContentType   string
	Headers       map[string]*Header
	Style         ParameterStyle
	Explode       bool
	AllowReserved bool
}
