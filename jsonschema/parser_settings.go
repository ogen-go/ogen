package jsonschema

import "errors"

// Settings is parser settings.
type Settings struct {
	External ExternalResolver
	// Resolver is a root resolver.
	Resolver ReferenceResolver

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
}
