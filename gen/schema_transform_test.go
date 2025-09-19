package gen

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/jsonpointer"
	"github.com/ogen-go/ogen/jsonschema"
)

func TestSingleOneOf(t *testing.T) {
	t.Run("single oneOf unwraps", func(t *testing.T) {
		a := require.New(t)
		s := createTestSchemaGen(nil)

		schema := &jsonschema.Schema{
			Type: jsonschema.Empty,
			OneOf: []*jsonschema.Schema{
				createPrimitiveSchema(jsonschema.String),
			},
		}
		result, err := s.generate("UnwrappedString", schema, false)
		a.NoError(err)
		a.NotNil(result)
		a.Equal(ir.KindPrimitive, result.Kind)
		a.Equal(ir.String, result.Primitive)
	})
}

func TestNullableOneOf_BasicPrimitives(t *testing.T) {
	primitiveTests := []struct {
		name         string
		schemaType   jsonschema.SchemaType
		format       string
		expectedType ir.PrimitiveType
	}{
		{"string", jsonschema.String, "", ir.String},
		{"integer", jsonschema.Integer, "", ir.Int},
		{"int32", jsonschema.Integer, "int32", ir.Int32},
		{"int64", jsonschema.Integer, "int64", ir.Int64},
		{"number", jsonschema.Number, "", ir.Float64},
		{"float", jsonschema.Number, "float", ir.Float32},
		{"double", jsonschema.Number, "double", ir.Float64},
		{"boolean", jsonschema.Boolean, "", ir.Bool},
	}

	for _, tc := range primitiveTests {
		t.Run(tc.name, func(t *testing.T) {
			a := require.New(t)
			s := createTestSchemaGen(nil)

			schema := createNullableOneOf(createPrimitiveSchema(tc.schemaType, tc.format))
			result, err := s.generate("Nullable"+tc.name, schema, false)

			a.NoError(err)
			a.NotNil(result)
			assertNullablePrimitive(t, result, tc.expectedType)
			assertNotSum(t, result)
		})
	}
}

func TestNullableOneOf_SpecialTypes(t *testing.T) {
	t.Run("byte string becomes nullable bytes", func(t *testing.T) {
		a := require.New(t)
		s := createTestSchemaGen(nil)

		schema := createNullableOneOf(createPrimitiveSchema(jsonschema.String, "byte"))
		result, err := s.generate("NullableBytes", schema, false)

		a.NoError(err)
		a.NotNil(result)
		a.Equal(ir.KindPrimitive, result.Kind)
		a.Equal(ir.ByteSlice, result.Primitive)
		a.Equal(ir.NilNull, result.NilSemantic)
		assertNotSum(t, result)
	})

	t.Run("array becomes nullable array", func(t *testing.T) {
		a := require.New(t)
		s := createTestSchemaGen(nil)

		arraySchema := &jsonschema.Schema{
			Type: jsonschema.Array,
			Item: createPrimitiveSchema(jsonschema.String),
		}
		schema := createNullableOneOf(arraySchema)
		result, err := s.generate("NullableArray", schema, false)

		a.NoError(err)
		a.NotNil(result)
		assertNullableArray(t, result)
		assertNotSum(t, result)
		// Verify item type
		a.Equal(ir.KindPrimitive, result.Item.Kind)
		a.Equal(ir.String, result.Item.Primitive)
	})

	t.Run("map becomes nullable map", func(t *testing.T) {
		a := require.New(t)
		s := createTestSchemaGen(nil)

		boolTrue := true
		mapSchema := &jsonschema.Schema{
			Type:                 jsonschema.Object,
			AdditionalProperties: &boolTrue,
			Item:                 createPrimitiveSchema(jsonschema.String),
		}
		schema := createNullableOneOf(mapSchema)
		result, err := s.generate("NullableMap", schema, false)

		a.NoError(err)
		a.NotNil(result)
		assertNullableGeneric(t, result, ir.KindMap)
		assertNotSum(t, result)
		// Verify value type
		a.Equal(ir.KindPrimitive, result.GenericOf.Item.Kind)
		a.Equal(ir.String, result.GenericOf.Item.Primitive)
	})
}

func TestNullableOneOf_Objects(t *testing.T) {
	t.Run("simple object", func(t *testing.T) {
		a := require.New(t)
		s := createTestSchemaGen(nil)

		objectSchema := createObjectSchema(
			createProperty("message", createPrimitiveSchema(jsonschema.String), true),
		)
		schema := createNullableOneOf(objectSchema)
		result, err := s.generate("NullableObject", schema, false)

		a.NoError(err)
		a.NotNil(result)
		assertNullableGeneric(t, result, ir.KindStruct)
		assertNotSum(t, result)
		// Verify object structure
		a.Len(result.GenericOf.Fields, 1)
		a.Equal("Message", result.GenericOf.Fields[0].Name)
	})

	t.Run("nested nullable objects", func(t *testing.T) {
		a := require.New(t)
		s := createTestSchemaGen(nil)

		innerObjectSchema := createObjectSchema(
			createProperty("description", createPrimitiveSchema(jsonschema.String), true),
		)
		outerObjectSchema := createObjectSchema(
			createProperty("name", createPrimitiveSchema(jsonschema.String), true),
			createProperty("details", createNullableOneOf(innerObjectSchema), true),
		)
		schema := createNullableOneOf(outerObjectSchema)
		result, err := s.generate("NullableNestedObject", schema, false)

		a.NoError(err)
		a.NotNil(result)
		assertNullableGeneric(t, result, ir.KindStruct)
		assertNotSum(t, result)

		// Verify nested structure
		structType := result.GenericOf
		a.Len(structType.Fields, 2)

		var detailsField *ir.Field
		for _, field := range structType.Fields {
			if field.Name == "Details" {
				detailsField = field
				break
			}
		}
		a.NotNil(detailsField, "Should have Details field")
		assertNotSum(t, detailsField.Type)
	})
}

func TestNullableOneOf_References(t *testing.T) {
	t.Run("single reference", func(t *testing.T) {
		a := require.New(t)

		resolver := func(ref jsonschema.Ref) (*ir.Type, bool) {
			if ref.Ptr == "#/components/schemas/User" {
				return &ir.Type{
					Name: "User",
					Kind: ir.KindStruct,
					Fields: []*ir.Field{
						{Name: "ID", Type: &ir.Type{Kind: ir.KindPrimitive, Primitive: ir.String}},
					},
				}, true
			}
			return nil, false
		}

		s := createTestSchemaGen(resolver)
		schema := &jsonschema.Schema{
			Type: jsonschema.Empty,
			OneOf: []*jsonschema.Schema{
				{Ref: jsonpointer.RefKey{Ptr: "#/components/schemas/User"}},
				{Type: jsonschema.Null},
			},
		}

		result, err := s.generate("NullableUser", schema, false)

		a.NoError(err)
		a.NotNil(result)
		assertNullableGeneric(t, result, ir.KindStruct)
		assertNotSum(t, result)
		a.Equal("User", result.GenericOf.Name)
	})

	t.Run("multiple references fail with discriminator error", func(t *testing.T) {
		a := require.New(t)

		resolver := func(ref jsonschema.Ref) (*ir.Type, bool) {
			switch ref.Ptr {
			case "#/components/schemas/User":
				return &ir.Type{Name: "User", Kind: ir.KindStruct, Fields: []*ir.Field{
					{Name: "ID", Type: &ir.Type{Kind: ir.KindPrimitive, Primitive: ir.String}},
				}}, true
			case "#/components/schemas/Admin":
				return &ir.Type{Name: "Admin", Kind: ir.KindStruct, Fields: []*ir.Field{
					{Name: "Role", Type: &ir.Type{Kind: ir.KindPrimitive, Primitive: ir.String}},
				}}, true
			}
			return nil, false
		}

		s := createTestSchemaGen(resolver)
		schema := &jsonschema.Schema{
			Type: jsonschema.Empty,
			OneOf: []*jsonschema.Schema{
				{Ref: jsonpointer.RefKey{Ptr: "#/components/schemas/User"}},
				{Ref: jsonpointer.RefKey{Ptr: "#/components/schemas/Admin"}},
				{Type: jsonschema.Null},
			},
		}

		result, err := s.generate("UserOrAdminOrNull", schema, false)

		a.Error(err)
		a.Contains(err.Error(), "discriminator inference not implemented")
		a.Nil(result)
	})
	t.Run("shared reference", func(t *testing.T) {
		a := require.New(t)
		s := createTestSchemaGen(nil)

		// Create a shared schema that will be referenced in multiple places
		sharedSchema := createPrimitiveSchema(jsonschema.String)

		// Create an object with two fields:
		// 1. One field with nullable oneOf (should be nullable)
		// 2. One field with direct reference to same schema (should NOT be nullable)
		objectSchema := createObjectSchema(
			createProperty("nullableField", createNullableOneOf(sharedSchema), true),
			createProperty("directField", sharedSchema, true),
		)

		result, err := s.generate("TestObject", objectSchema, false)
		a.NoError(err)
		a.NotNil(result)
		a.Equal(ir.KindStruct, result.Kind)
		a.Len(result.Fields, 2)

		var nullableField, directField *ir.Field
		for _, field := range result.Fields {
			switch field.Name {
			case "NullableField":
				nullableField = field
			case "DirectField":
				directField = field
			}
		}

		// Nullable field should be nullable
		a.NotNil(nullableField)
		assertNullablePrimitive(t, nullableField.Type, ir.String)

		// Direct field should NOT be nullable
		a.NotNil(directField)
		a.Equal(ir.KindPrimitive, directField.Type.Kind, "Direct field should be primitive, not nullable")
		a.Equal(ir.String, directField.Type.Primitive, "Direct field should be string")
	})
}

func TestNullableOneOf_Optionality(t *testing.T) {
	t.Run("required vs optional nullable fields", func(t *testing.T) {
		a := require.New(t)
		s := createTestSchemaGen(nil)

		nullableStringSchema := createNullableOneOf(createPrimitiveSchema(jsonschema.String))

		// Test required nullable field
		requiredSchema := createObjectSchema(
			createProperty("requiredNullable", nullableStringSchema, true),
		)
		requiredResult, err := s.generate("WithRequired", requiredSchema, false)

		a.NoError(err)
		a.NotNil(requiredResult)
		a.Equal(ir.KindStruct, requiredResult.Kind)
		a.Len(requiredResult.Fields, 1)

		requiredField := requiredResult.Fields[0]
		a.Equal("RequiredNullable", requiredField.Name)
		// Required nullable: only nullable flag
		a.Equal(ir.KindGeneric, requiredField.Type.Kind)
		a.True(requiredField.Type.GenericVariant.Nullable)
		a.False(requiredField.Type.GenericVariant.Optional)

		// Test optional nullable field
		optionalSchema := createObjectSchema(
			createProperty("optionalNullable", nullableStringSchema, false),
		)
		optionalResult, err := s.generate("WithOptional", optionalSchema, false)

		a.NoError(err)
		a.NotNil(optionalResult)
		a.Equal(ir.KindStruct, optionalResult.Kind)
		a.Len(optionalResult.Fields, 1)

		optionalField := optionalResult.Fields[0]
		a.Equal("OptionalNullable", optionalField.Name)
		// Optional nullable: both flags
		assertOptionalNullableGeneric(t, optionalField.Type, ir.KindPrimitive)
	})

	t.Run("field in optional object", func(t *testing.T) {
		a := require.New(t)
		s := createTestSchemaGen(nil)

		nullableStringSchema := createNullableOneOf(createPrimitiveSchema(jsonschema.String))
		objectSchema := createObjectSchema(
			createProperty("message", nullableStringSchema, false),
		)

		// Generate with optional=true at the object level
		result, err := s.generate("OptionalObject", objectSchema, true)

		a.NoError(err)
		a.NotNil(result)
		// The object itself should be optional
		a.Equal(ir.KindGeneric, result.Kind)
		a.True(result.GenericVariant.Optional)
		a.False(result.GenericVariant.Nullable)

		// The field inside should still be optional+nullable
		innerStruct := result.GenericOf
		a.Equal(ir.KindStruct, innerStruct.Kind)
		a.Len(innerStruct.Fields, 1)

		field := innerStruct.Fields[0]
		assertOptionalNullableGeneric(t, field.Type, ir.KindPrimitive)
	})
}

func TestNullableOneOf_NonNullablePatterns(t *testing.T) {
	t.Run("multiple non-null types remain sum type", func(t *testing.T) {
		nonNullTests := []struct {
			name     string
			schema   *jsonschema.Schema
			expected int
		}{
			{
				"string or integer",
				&jsonschema.Schema{
					Type: jsonschema.Empty,
					OneOf: []*jsonschema.Schema{
						createPrimitiveSchema(jsonschema.String),
						createPrimitiveSchema(jsonschema.Integer),
					},
				},
				2,
			},
			{
				"string, integer, and null",
				&jsonschema.Schema{
					Type: jsonschema.Empty,
					OneOf: []*jsonschema.Schema{
						createPrimitiveSchema(jsonschema.String),
						createPrimitiveSchema(jsonschema.Integer),
						{Type: jsonschema.Null},
					},
				},
				3,
			},
			{
				"string, integer, and object",
				&jsonschema.Schema{
					Type: jsonschema.Empty,
					OneOf: []*jsonschema.Schema{
						createPrimitiveSchema(jsonschema.String),
						createPrimitiveSchema(jsonschema.Integer),
						createObjectSchema(
							createProperty("value", createPrimitiveSchema(jsonschema.String), true),
						),
					},
				},
				3,
			},
		}

		for _, tc := range nonNullTests {
			t.Run(tc.name, func(t *testing.T) {
				a := require.New(t)
				s := createTestSchemaGen(nil)

				result, err := s.generate("MultiType", tc.schema, false)

				a.NoError(err)
				a.NotNil(result)
				assertSumType(t, result, tc.expected)
			})
		}
	})
}

func createTestSchemaGen(resolver func(ref jsonschema.Ref) (*ir.Type, bool)) *schemaGen {
	if resolver == nil {
		resolver = func(ref jsonschema.Ref) (*ir.Type, bool) { return nil, false }
	}
	core, _ := observer.New(zap.InfoLevel)
	s := newSchemaGen(resolver)
	s.log = zap.New(core)
	return s
}

func createNullableOneOf(typeSchema *jsonschema.Schema) *jsonschema.Schema {
	return &jsonschema.Schema{
		Type: jsonschema.Empty,
		OneOf: []*jsonschema.Schema{
			typeSchema,
			{Type: jsonschema.Null},
		},
	}
}

func createPrimitiveSchema(typ jsonschema.SchemaType, format ...string) *jsonschema.Schema {
	s := &jsonschema.Schema{Type: typ}
	if len(format) > 0 {
		s.Format = format[0]
	}
	return s
}

func createObjectSchema(properties ...jsonschema.Property) *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:       jsonschema.Object,
		Properties: properties,
	}
}

func createProperty(name string, schema *jsonschema.Schema, required bool) jsonschema.Property {
	return jsonschema.Property{
		Name:     name,
		Schema:   schema,
		Required: required,
	}
}

// Assertion helpers for common test patterns

func assertNullableGeneric(t *testing.T, result *ir.Type, expectedInnerKind ir.Kind) {
	t.Helper()
	a := require.New(t)
	a.Equal(ir.KindGeneric, result.Kind, "Expected generic wrapper")
	a.Equal(expectedInnerKind, result.GenericOf.Kind, "Unexpected inner type")
	a.True(result.GenericVariant.Nullable, "Expected nullable variant")
	a.False(result.GenericVariant.Optional, "Expected non-optional at type level")
}

func assertNullablePrimitive(t *testing.T, result *ir.Type, expectedPrimitive ir.PrimitiveType) {
	t.Helper()
	assertNullableGeneric(t, result, ir.KindPrimitive)
	a := require.New(t)
	a.Equal(expectedPrimitive, result.GenericOf.Primitive, "Unexpected primitive type")
}

func assertOptionalNullableGeneric(t *testing.T, result *ir.Type, expectedInnerKind ir.Kind) {
	t.Helper()
	a := require.New(t)
	a.Equal(ir.KindGeneric, result.Kind, "Expected generic wrapper")
	a.Equal(expectedInnerKind, result.GenericOf.Kind, "Unexpected inner type")
	a.True(result.GenericVariant.Nullable, "Expected nullable variant")
	a.True(result.GenericVariant.Optional, "Expected optional variant")
}

func assertNullableArray(t *testing.T, result *ir.Type) {
	t.Helper()
	a := require.New(t)
	a.Equal(ir.KindArray, result.Kind, "Expected array type")
	a.Equal(ir.NilNull, result.NilSemantic, "Expected NilNull semantic for nullable array")
}

func assertSumType(t *testing.T, result *ir.Type, expectedVariants int) {
	t.Helper()
	a := require.New(t)
	a.Equal(ir.KindSum, result.Kind, "Expected sum type")
	a.Len(result.SumOf, expectedVariants, "Unexpected number of sum variants")
}

func assertNotSum(t *testing.T, result *ir.Type) {
	t.Helper()
	a := require.New(t)
	a.NotEqual(ir.KindSum, result.Kind, "Expected non-sum type")
}
