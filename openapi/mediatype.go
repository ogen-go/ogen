package openapi

import (
	"encoding/json"

	"github.com/ogen-go/ogen/jsonschema"
)

type MediaType struct {
	Schema   *jsonschema.Schema
	Example  json.RawMessage
	Examples map[string]*Example
	// Encoding map[string]*Encoding
}
