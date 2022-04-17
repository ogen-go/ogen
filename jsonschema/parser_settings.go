package jsonschema

import "errors"

type Settings struct {
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
	if s.Resolver == nil {
		s.Resolver = nopResolver{}
	}
}
