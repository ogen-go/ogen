package parser

import (
	"net/url"
	"strings"

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

	// AuthenticationSchemes is the list of allowed HTTP Authorization schemes in a Security Scheme Object.
	//
	// Authorization schemes are case-insensitive.
	//
	// If empty, the ones registered in https://www.iana.org/assignments/http-authschemes/http-authschemes.xhtml are used.
	//
	// See https://swagger.io/specification/#security-scheme-object.
	AuthenticationSchemes []string
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
	if len(s.AuthenticationSchemes) != 0 {
		// Make sure schemes are lowercased
		for i, scheme := range s.AuthenticationSchemes {
			s.AuthenticationSchemes[i] = strings.ToLower(scheme)
		}
	} else {
		// Values from https://www.iana.org/assignments/http-authschemes/http-authschemes.xhtml.
		s.AuthenticationSchemes = []string{
			"basic",
			"bearer",
			"concealed",
			"digest",
			"dpop",
			"gnap",
			"hoba",
			"mutual",
			"negotiate",
			"oauth",
			"privatetoken",
			"scram-sha-1",
			"scram-sha-256",
			"vapid",
		}
	}
}
