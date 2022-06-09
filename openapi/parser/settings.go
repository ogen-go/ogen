package parser

import "github.com/ogen-go/ogen/jsonschema"

// Settings is parser settings.
type Settings struct {
	External jsonschema.ExternalResolver

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
