package openapi

import (
	"encoding/json"

	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/location"
)

// SSEEventShape describes how an SSE event is represented by a schema.
type SSEEventShape string

const (
	// SSEEventShapeNone means media type is not generated as SSE.
	SSEEventShapeNone SSEEventShape = ""
	// SSEEventShapeDataOnly means schema describes only the SSE data payload.
	SSEEventShapeDataOnly SSEEventShape = "data-only"
	// SSEEventShapeFull means schema describes the full SSE event envelope.
	SSEEventShapeFull SSEEventShape = "full"
	// SSEEventShapeFullArray means schema describes the SSE stream as an array
	// of full SSE event envelopes.
	SSEEventShapeFullArray SSEEventShape = "full-array"
)

// Enabled returns true if the media type should be generated as SSE.
func (s SSEEventShape) Enabled() bool { return s != SSEEventShapeNone }

// MediaType is Media Type Object.
type MediaType struct {
	Schema   *jsonschema.Schema
	Example  json.RawMessage
	Examples map[string]*Example
	Encoding map[string]*Encoding

	XOgenJSONStreaming bool
	XOgenRawResponse   bool
	XOgenSSEEventShape SSEEventShape

	location.Pointer `json:"-" yaml:"-"`
}

// Encoding is Encoding Type Object.
type Encoding struct {
	ContentType   string
	Headers       map[string]*Header
	Style         ParameterStyle
	Explode       bool
	AllowReserved bool

	location.Pointer `json:"-" yaml:"-"`
}
