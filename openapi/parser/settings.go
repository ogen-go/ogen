package parser

import (
	"net/url"

	"github.com/ogen-go/ogen/jsonpointer"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/location"
)

// Settings is parser settings.
type Settings struct {
	// External is external JSON Schema resolver. If nil, NoExternal resolver is used.
	External jsonschema.ExternalResolver

	// File is the file that is being parsed.
	//
	// Used for error messages.
	File location.File

	// RootURL is the root URL of the spec.
	//
	// If nil, jsonpointer.DummyURL is used.
	RootURL *url.URL

	// DepthLimit limits the number of nested references. Default is 1000.
	DepthLimit int

	// Enables type inference.
	//
	// For example:
	//
	//	{
	//		"items": {
	//			"type": "string"
	//		}
	//	}
	//
	// In that case schemaParser will handle that schema as "array" schema, because it has "items" field.
	InferTypes bool
}

func (s *Settings) setDefaults() {
	if s.External == nil {
		s.External = jsonschema.NoExternal{}
	}
	if s.DepthLimit == 0 {
		s.DepthLimit = jsonpointer.DefaultDepthLimit
	}
	if s.RootURL == nil {
		s.RootURL = jsonpointer.DummyURL()
	}
}
