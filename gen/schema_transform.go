package gen

import (
	"github.com/ogen-go/ogen/jsonschema"
)

func transformSchema(schema *jsonschema.Schema) *jsonschema.Schema {
	if schema == nil {
		return nil
	}
	schema = transformSingleOneOf(schema)
	schema = transformNullableOneOf(schema)
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

// transformNullableOneOf detects and handles nullable oneOf patterns.
//
// nullable oneOf pattern is:
//
//	oneOf:
//	  - type: "null"
//	  - <schema>
//
// if such pattern is detected, this function will return a Nulllable version of the inner schema.
func transformNullableOneOf(schema *jsonschema.Schema) *jsonschema.Schema {
	if schema == nil || len(schema.OneOf) != 2 {
		return schema
	}

	var nullSchema, nonNullSchema *jsonschema.Schema
	for _, s := range schema.OneOf {
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
