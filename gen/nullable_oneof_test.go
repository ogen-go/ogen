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

func TestOneOfWithNullBecomesNullable(t *testing.T) {
	t.Run("oneOf with string and null becomes nullable string", func(t *testing.T) {
		a := require.New(t)

		core, _ := observer.New(zap.InfoLevel)
		s := newSchemaGen(func(ref jsonschema.Ref) (*ir.Type, bool) { return nil, false })
		s.log = zap.New(core)

		schema := &jsonschema.Schema{
			Type: jsonschema.Empty,
			OneOf: []*jsonschema.Schema{
				{
					Type: jsonschema.String,
				},
				{
					Type: jsonschema.Null,
				},
			},
		}

		result, err := s.oneOf("NullableString", schema, false)

		a.NoError(err)
		a.NotNil(result)

		// Should be a nullable primitive string wrapped in a generic, not a sum type
		a.Equal(ir.KindGeneric, result.Kind)
		a.Equal(ir.KindPrimitive, result.GenericOf.Kind)
		a.Equal(ir.String, result.GenericOf.Primitive)
		a.True(result.GenericVariant.Nullable, "Result should be nullable")
		a.False(result.IsSum(), "Result should not be a sum type")
	})

	t.Run("oneOf with object and null becomes nullable object", func(t *testing.T) {
		a := require.New(t)

		core, _ := observer.New(zap.InfoLevel)
		s := newSchemaGen(func(ref jsonschema.Ref) (*ir.Type, bool) { return nil, false })
		s.log = zap.New(core)

		schema := &jsonschema.Schema{
			Type: jsonschema.Empty,
			OneOf: []*jsonschema.Schema{
				{
					Type: jsonschema.Object,
					Properties: []jsonschema.Property{
						{
							Name: "message",
							Schema: &jsonschema.Schema{
								Type: jsonschema.String,
							},
						},
					},
				},
				{
					Type: jsonschema.Null,
				},
			},
		}

		result, err := s.oneOf("NullableObject", schema, false)

		a.NoError(err)
		a.NotNil(result)

		// Should be a nullable struct wrapped in a generic, not a sum type
		a.Equal(ir.KindGeneric, result.Kind)
		a.Equal(ir.KindStruct, result.GenericOf.Kind)
		a.True(result.GenericVariant.Nullable, "Result should be nullable")
		a.False(result.IsSum(), "Result should not be a sum type")
		a.Len(result.GenericOf.Fields, 1)
		a.Equal("Message", result.GenericOf.Fields[0].Name)
	})

	t.Run("oneOf with reference and null becomes nullable reference", func(t *testing.T) {
		a := require.New(t)

		// Mock resolver
		resolver := func(ref jsonschema.Ref) (*ir.Type, bool) {
			if ref.Ptr == "#/components/schemas/Error" {
				return &ir.Type{
					Name: "Error",
					Kind: ir.KindStruct,
					Fields: []*ir.Field{
						{
							Name: "Message",
							Type: &ir.Type{Kind: ir.KindPrimitive, Primitive: ir.String},
							Tag:  ir.Tag{JSON: "message"},
						},
					},
				}, true
			}
			return nil, false
		}

		core, _ := observer.New(zap.InfoLevel)
		s := newSchemaGen(resolver)
		s.log = zap.New(core)

		schema := &jsonschema.Schema{
			Type: jsonschema.Empty,
			OneOf: []*jsonschema.Schema{
				{
					Ref: jsonpointer.RefKey{Ptr: "#/components/schemas/Error"},
				},
				{
					Type: jsonschema.Null,
				},
			},
		}

		result, err := s.oneOf("NullableError", schema, false)

		a.NoError(err)
		a.NotNil(result)

		// Should be a nullable reference wrapped in a generic, not a sum type
		// For references, the name might be different based on how the resolver works
		a.Equal(ir.KindGeneric, result.Kind)
		a.True(result.GenericVariant.Nullable, "Result should be nullable")
		a.False(result.IsSum(), "Result should not be a sum type")

		// The underlying type should be the resolved reference
		a.Equal(ir.KindStruct, result.GenericOf.Kind)
	})

	t.Run("oneOf with multiple non-null types should remain sum type", func(t *testing.T) {
		a := require.New(t)

		core, _ := observer.New(zap.InfoLevel)
		s := newSchemaGen(func(ref jsonschema.Ref) (*ir.Type, bool) { return nil, false })
		s.log = zap.New(core)

		schema := &jsonschema.Schema{
			Type: jsonschema.Empty,
			OneOf: []*jsonschema.Schema{
				{
					Type: jsonschema.String,
				},
				{
					Type: jsonschema.Integer,
				},
			},
		}

		result, err := s.oneOf("StringOrInt", schema, false)

		a.NoError(err)
		a.NotNil(result)

		// Should remain a sum type since no null involved
		a.Equal(ir.KindSum, result.Kind)
		a.True(result.IsSum())
		a.Len(result.SumOf, 2)
	})

	t.Run("oneOf with string, integer, and null should remain the same type", func(t *testing.T) {
		a := require.New(t)

		core, _ := observer.New(zap.InfoLevel)
		s := newSchemaGen(func(ref jsonschema.Ref) (*ir.Type, bool) { return nil, false })
		s.log = zap.New(core)

		// This has 3 variants including null, so it's not the simple nullable pattern
		schema := &jsonschema.Schema{
			Type: jsonschema.Empty,
			OneOf: []*jsonschema.Schema{
				{
					Type: jsonschema.String,
				},
				{
					Type: jsonschema.Integer,
				},
				{
					Type: jsonschema.Null,
				},
			},
		}

		result, err := s.oneOf("StringOrIntOrNull", schema, false)

		a.NoError(err)
		a.NotNil(result)

		// Should be a sum type since we have multiple non-null types
		a.Equal(ir.KindSum, result.Kind)
		a.True(result.IsSum())
		a.Len(result.SumOf, 3)
	})

	t.Run("oneOf with string, integer, and object should remain the same type", func(t *testing.T) {
		a := require.New(t)

		core, _ := observer.New(zap.InfoLevel)
		s := newSchemaGen(func(ref jsonschema.Ref) (*ir.Type, bool) { return nil, false })
		s.log = zap.New(core)

		// This has 3 variants including null, so it's not the simple nullable pattern
		schema := &jsonschema.Schema{
			Type: jsonschema.Empty,
			OneOf: []*jsonschema.Schema{
				{
					Type: jsonschema.String,
				},
				{
					Type: jsonschema.Integer,
				},
				{
					Type: jsonschema.Object,
					Properties: []jsonschema.Property{
						{
							Name: "message",
							Schema: &jsonschema.Schema{
								Type: jsonschema.String,
							},
						},
					},
				},
			},
		}

		result, err := s.oneOf("StringOrIntOrObject", schema, false)

		a.NoError(err)
		a.NotNil(result)

		// Should be a sum type since we have multiple non-null types
		a.Equal(ir.KindSum, result.Kind)
		a.True(result.IsSum())
		a.Len(result.SumOf, 3)
	})
}
