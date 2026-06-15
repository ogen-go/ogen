package gen

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/jsonschema"
)

func Test_mergeEnums(t *testing.T) {
	tests := []struct {
		a, b    []any
		want    []any
		wantErr bool
	}{
		// Fast path.
		{nil, nil, nil, false},
		{[]any{1, 2, 3}, nil, []any{1, 2, 3}, false},
		// Merge.
		{[]any{1, 2, 3}, []any{3, 4, 5}, []any{3}, false},
		{[]any{3}, []any{3, 4, 5}, []any{3}, false},
		{[]any{"a"}, []any{"b", "a", "c"}, []any{"a"}, false},
		{[]any{
			"a", "b",
			0, 2,
			false, true,
			[]any{1},
			[]any{2},
		}, []any{
			"a", "c",
			0, 3,
			true,
			[]any{1},
			[]any{[]any{1}},
		}, []any{
			"a",
			0,
			true,
			[]any{1},
		}, false},
		// No common values.
		{[]any{1, 2, 3}, []any{4, 5, 6}, nil, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			// Ensure that merge is commutative.
			got1, err1 := mergeEnums(
				&jsonschema.Schema{Enum: tt.a},
				&jsonschema.Schema{Enum: tt.b},
			)
			got2, err2 := mergeEnums(
				&jsonschema.Schema{Enum: tt.b},
				&jsonschema.Schema{Enum: tt.a},
			)
			if tt.wantErr {
				a.Error(err1)
				a.Error(err2)
				return
			}
			a.NoError(err1)
			a.NoError(err2)
			a.Equal(tt.want, got1)
			a.Equal(tt.want, got2)
		})
	}
}

// TestAllOfWithSiblingProperties ensures sibling keywords specified alongside
// allOf (properties, required, ...) are merged in rather than dropped.
//
// Per JSON Schema, allOf and its sibling keywords are all applied (logical AND).
func TestAllOfWithSiblingProperties(t *testing.T) {
	t.Run("single allOf subschema", func(t *testing.T) {
		a := require.New(t)
		s := createTestSchemaGen(nil)

		// type: object
		// allOf:
		//   - type: object
		//     properties: { fromAllOf: { type: string } }
		// required: [sibling]
		// properties:
		//   sibling: { type: string }
		schema := &jsonschema.Schema{
			Type: jsonschema.Object,
			AllOf: []*jsonschema.Schema{
				createObjectSchema(
					createProperty("fromAllOf", createPrimitiveSchema(jsonschema.String), false),
				),
			},
			Required: []string{"sibling"},
			Properties: []jsonschema.Property{
				createProperty("sibling", createPrimitiveSchema(jsonschema.String), true),
			},
		}

		result, err := s.generate("AllOfWithSibling", schema, false)
		a.NoError(err)
		a.Equal(ir.KindStruct, result.Kind)

		fields := map[string]*ir.Field{}
		for _, f := range result.Fields {
			fields[f.Name] = f
		}
		a.Contains(fields, "FromAllOf")
		a.Contains(fields, "Sibling", "sibling property must not be dropped when allOf is present")
	})

	t.Run("multiple allOf subschemas", func(t *testing.T) {
		a := require.New(t)
		s := createTestSchemaGen(nil)

		schema := &jsonschema.Schema{
			Type: jsonschema.Object,
			AllOf: []*jsonschema.Schema{
				createObjectSchema(createProperty("a", createPrimitiveSchema(jsonschema.String), false)),
				createObjectSchema(createProperty("b", createPrimitiveSchema(jsonschema.String), false)),
			},
			Properties: []jsonschema.Property{
				createProperty("sibling", createPrimitiveSchema(jsonschema.String), false),
			},
		}

		result, err := s.generate("MultiAllOfWithSibling", schema, false)
		a.NoError(err)
		a.Equal(ir.KindStruct, result.Kind)

		var names []string
		for _, f := range result.Fields {
			names = append(names, f.Name)
		}
		a.ElementsMatch([]string{"A", "B", "Sibling"}, names)
	})
}

// TestAllOfPropagatesParentExtensions ensures ogen-specific keywords specified
// alongside allOf (x-ogen-validate, x-oapi-codegen-extra-tags,
// x-ogen-time-format, xml) survive flattening instead of being dropped.
func TestAllOfPropagatesParentExtensions(t *testing.T) {
	a := require.New(t)

	parent := &jsonschema.Schema{
		Type:            jsonschema.Object,
		OgenValidate:    map[string]any{"required": true},
		ExtraTags:       map[string]string{"validate": "required"},
		XOgenTimeFormat: "unix",
		XML:             &jsonschema.XML{Name: "pet"},
		AllOf: []*jsonschema.Schema{
			createObjectSchema(
				createProperty("fromAllOf", createPrimitiveSchema(jsonschema.String), false),
			),
		},
	}

	got, err := flattenAllOfSchema(parent)
	a.NoError(err)
	// The single-allOf shortcut must be skipped so the extensions can be applied
	// (a referenced/ref-bearing result would bypass the post-flatten consumers).
	a.True(got.Ref.IsZero())
	a.Equal(map[string]any{"required": true}, got.OgenValidate)
	a.Equal(map[string]string{"validate": "required"}, got.ExtraTags)
	a.Equal("unix", got.XOgenTimeFormat)
	a.NotNil(got.XML)
	a.Equal("pet", got.XML.Name)
}
