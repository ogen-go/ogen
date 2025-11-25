package jsonschema

import (
	"errors"

	"github.com/ogen-go/ogen/location"
)

// Settings is parser settings.
type Settings struct {
	// External is external resolver. If nil, NoExternal resolver is used.
	External ExternalResolver

	// Resolver is a root resolver.
	Resolver ReferenceResolver

	// File is the file that is being parsed.
	//
	// Used for error messages.
	File location.File

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

	// AllowCrossTypeConstraints enables interpretation of cross-type schema constraints.
	// When true (default), constraints like pattern on numbers or maximum on strings
	// are allowed and interpreted during code generation.
	AllowCrossTypeConstraints bool
}

type nopResolver struct{}

func (nopResolver) ResolveReference(ref string) (*RawSchema, error) {
	return nil, errors.New("reference resolver is not provided")
}

func (s *Settings) setDefaults() {
	if s.External == nil {
		s.External = NoExternal{}
	}
	if s.Resolver == nil {
		s.Resolver = nopResolver{}
	}
	// Default: allow cross-type constraints (interpret them correctly)
	// This can be overridden by explicitly setting to false
}
