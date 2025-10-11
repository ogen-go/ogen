package gen

import (
	"github.com/ogen-go/ogen/jsonschema"
)

func transformSchema(schema *jsonschema.Schema) *jsonschema.Schema {
	if schema == nil {
		return nil
	}
	schema = transformSingleOneOf(schema)
	schema = transformNullableUnionType(schema)
	return schema
}

// transformSingleOneOf detects and handles single oneOf patterns.
//
// single oneOf pattern is:
//
//	oneOf:
//	  - <schema>
//
// if such pattern is detected, this function will return the inner schema.
func transformSingleOneOf(schema *jsonschema.Schema) *jsonschema.Schema {
	if schema.Discriminator == nil && len(schema.OneOf) == 1 {
		return schema.OneOf[0]
	}
	return schema
}

// transformNullableUnionType detects and handles nullable oneOf/anyOf patterns.
//
// nullable oneOf pattern is:
//
//	oneOf:
//	  - type: "null"
//	  - <schema>
//
// or
//
//	anyOf:
//	  - type: "null"
//	  - <schema>
//
// if such pattern is detected, this function will return a Nulllable version of the inner schema.
func transformNullableUnionType(schema *jsonschema.Schema) *jsonschema.Schema {
	if schema == nil {
		return nil
	}

	var schemas []*jsonschema.Schema

	switch {
	case len(schema.AnyOf) == 2:
		schemas = schema.AnyOf
	case len(schema.OneOf) == 2:
		schemas = schema.OneOf
	default:
		return schema
	}

	var nullSchema, nonNullSchema *jsonschema.Schema
	for _, s := range schemas {
		if s != nil {
			if s.Type == jsonschema.Null {
				nullSchema = s
			} else {
				nonNullSchema = s
			}
		}
	}

	// If we didn't find exactly one null and one non-null variant, don't handle
	if nullSchema == nil || nonNullSchema == nil {
		return schema
	}

	// Return nullable version of the underlined schema.
	// Make a shallow copy to avoid mutating the original schema.
	nullableSchema := *nonNullSchema
	nullableSchema.Nullable = true
	return &nullableSchema
}
