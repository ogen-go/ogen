package gen

import (
	"fmt"
	"testing"

	"github.com/go-faster/yaml"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/jsonpointer"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/location"
)

func TestSchemaGenAnyWarn(t *testing.T) {
	a := require.New(t)

	core, ob := observer.New(zap.InfoLevel)
	s := newSchemaGen(func(ref jsonschema.Ref) (*ir.Type, bool) {
		return nil, false
	})
	s.log = zap.New(core)

	_, err := s.generate("foo", &jsonschema.Schema{
		Type: "",
	}, false)
	a.NoError(err)

	entries := ob.FilterMessage("Type is not defined, using any").All()
	a.Len(entries, 1)
	args := entries[0].ContextMap()
	a.Equal("foo", args["name"])
}

func TestSchemaGenNilSchema(t *testing.T) {
	a := require.New(t)

	t.Run("Response", func(t *testing.T) {
		s := newSchemaGen(func(ref jsonschema.Ref) (*ir.Type, bool) {
			return nil, false
		})
		s.request = false // response

		// Test that nil schema in responses is handled as "any" (jx.Raw).
		// This occurs when response has content without a schema field,
		// e.g., default error responses: {"content": {"application/json": {}}}
		typ, err := s.generate("test", nil, false)
		a.NoError(err)
		a.NotNil(typ)
		a.Equal(ir.KindAny, typ.Kind)
		a.Equal("jx.Raw", typ.Go())
	})

	t.Run("Request", func(t *testing.T) {
		s := newSchemaGen(func(ref jsonschema.Ref) (*ir.Type, bool) {
			return nil, false
		})
		s.request = true // request

		// Test that nil schema in requests returns an error.
		// Clients shouldn't send arbitrary data without explicit schema guidance.
		_, err := s.generate("test", nil, false)
		a.Error(err)
		a.Contains(err.Error(), "empty schema in request body")
	})
}

func TestSchemaGenConst(t *testing.T) {
	a := require.New(t)

	tests := []struct {
		name          string
		schema        *jsonschema.Schema
		expectedConst any
	}{
		{
			name: "integer const",
			schema: &jsonschema.Schema{
				Type: jsonschema.Object,
				Properties: []jsonschema.Property{
					{
						Name: "code",
						Schema: &jsonschema.Schema{
							Type:     jsonschema.Integer,
							Const:    int64(400),
							ConstSet: true,
						},
						Required: true,
					},
				},
			},
			expectedConst: int64(400),
		},
		{
			name: "string const",
			schema: &jsonschema.Schema{
				Type: jsonschema.Object,
				Properties: []jsonschema.Property{
					{
						Name: "status",
						Schema: &jsonschema.Schema{
							Type:     jsonschema.String,
							Const:    "active",
							ConstSet: true,
						},
						Required: true,
					},
				},
			},
			expectedConst: "active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newSchemaGen(func(ref jsonschema.Ref) (*ir.Type, bool) {
				return nil, false
			})

			typ, err := s.generate("Test", tt.schema, false)
			a.NoError(err)
			a.NotNil(typ)
			a.Equal(ir.KindStruct, typ.Kind)
			a.Len(typ.Fields, 1)

			field := typ.Fields[0]
			constVal := field.Const()
			a.True(constVal.Set, "field should have Const().Set=true")
			a.Equal(tt.expectedConst, constVal.Value, "const value should match")
			a.Equal(ir.KindPrimitive, field.Type.Kind, "field type should be primitive, not enum")
		})
	}
}

func TestGenerate(t *testing.T) {
	var loc location.Locator
	loc.UnmarshalYAML(&yaml.Node{
		Line:   1,
		Column: 1,
	})
	pointer := location.Pointer{
		Source: location.File{
			Name:   "pet-tags.yml",
			Source: "pet-tags.yml",
		},
		Locator: loc,
	}

	tests := []struct {
		name           string
		schema         *jsonschema.Schema
		optional       bool
		expectedIrType func(*jsonschema.Schema) *ir.Type
		expectedErr    string
	}{
		{
			name: "fieldTag",
			schema: &jsonschema.Schema{
				Ref: jsonpointer.RefKey{
					Loc: "jsonschema://dummy",
					Ptr: "#/components/schemas/Pet",
				},
				Type: "object",
				Properties: []jsonschema.Property{
					{
						Name: "id",
						Schema: &jsonschema.Schema{
							Type:   "integer",
							Format: "int64",
							ExtraTags: map[string]string{
								"gorm":  "primaryKey",
								"valid": "customIdValidator",
							},
							Pointer: pointer,
						},
						Required: true,
					},

					{
						Name: "name",
						Schema: &jsonschema.Schema{
							Type: "string",
							ExtraTags: map[string]string{
								"valid": "customNameValidator",
							},
							Pointer: pointer,
						},
						Required: true,
					},

					{
						Name: "tag",
						Schema: &jsonschema.Schema{
							Type:    "string",
							Pointer: pointer,
						},
						Required: true,
					},
				},
			},
			optional:    false,
			expectedErr: "",
			expectedIrType: func(schema *jsonschema.Schema) *ir.Type {
				return &ir.Type{
					Kind: "struct",
					Name: "Pet",
					Fields: []*ir.Field{
						{
							Name: "ID",
							Type: &ir.Type{
								Kind:      "primitive",
								Primitive: ir.Int64,
								Schema:    schema.Properties[0].Schema,
							},
							Tag: ir.Tag{
								JSON: "id",
								ExtraTags: map[string]string{
									"gorm":  "primaryKey",
									"valid": "customIdValidator",
								},
							},
							Spec: &jsonschema.Property{
								Name:     "id",
								Schema:   schema.Properties[0].Schema,
								Required: true,
							},
						},

						{
							Name: "Name",
							Type: &ir.Type{
								Kind:      "primitive",
								Primitive: ir.String,
								Schema:    schema.Properties[1].Schema,
							},
							Tag: ir.Tag{
								JSON: "name",
								ExtraTags: map[string]string{
									"valid": "customNameValidator",
								},
							},
							Spec: &jsonschema.Property{
								Name:     "name",
								Schema:   schema.Properties[1].Schema,
								Required: true,
							},
						},

						{
							Name: "Tag",
							Type: &ir.Type{
								Kind:      "primitive",
								Primitive: ir.String,
								Schema:    schema.Properties[2].Schema,
							},
							Tag: ir.Tag{
								JSON: "tag",
							},
							Spec: &jsonschema.Property{
								Name:     "tag",
								Schema:   schema.Properties[2].Schema,
								Required: true,
							},
						},
					},
					Schema: schema,
				}
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test %s", tt.name), func(t *testing.T) {
			a := require.New(t)

			core, _ := observer.New(zap.InfoLevel)
			s := newSchemaGen(func(ref jsonschema.Ref) (*ir.Type, bool) {
				return nil, false
			})
			s.log = zap.New(core)

			irType, err := s.generate(tt.name, tt.schema, tt.optional)

			var errText string
			if err != nil {
				errText = err.Error()
			}
			a.Equal(tt.expectedErr, errText, "err")

			expectedIrType := tt.expectedIrType(tt.schema)
			expectedIrTypeY, _ := yaml.Marshal(expectedIrType)
			irTypeY, _ := yaml.Marshal(irType)
			a.Equal(expectedIrType, irType, fmt.Sprintf("\nEXPECTED:\n\n%s\nACTUAL:\n\n%s", expectedIrTypeY, irTypeY))
		})
	}
}
