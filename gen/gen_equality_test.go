package gen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/gen/ir"
)

func TestCreateEqualityMethodSpec_ArrayDetection(t *testing.T) {
	tests := []struct {
		name           string
		irType         *ir.Type
		expectedFields map[string]fieldAssertions
	}{
		{
			name: "optional array field",
			irType: &ir.Type{
				Name: "TestType",
				Kind: ir.KindStruct,
				Fields: []*ir.Field{
					{
						Name: "OptionalArray",
						Type: &ir.Type{
							Name: "OptStringArray",
							Kind: ir.KindGeneric,
							GenericOf: &ir.Type{
								Kind: ir.KindArray,
								Item: &ir.Type{
									Kind:      ir.KindPrimitive,
									Primitive: ir.String,
								},
							},
						},
					},
				},
			},
			expectedFields: map[string]fieldAssertions{
				"OptionalArray": {
					fieldType: ir.FieldTypeOptional,
					isArray:   true,
				},
			},
		},
		{
			name: "nullable array field",
			irType: &ir.Type{
				Name: "TestType",
				Kind: ir.KindStruct,
				Fields: []*ir.Field{
					{
						Name: "NullableArray",
						Type: &ir.Type{
							Name: "NilIntArray",
							Kind: ir.KindGeneric,
							GenericOf: &ir.Type{
								Kind: ir.KindArray,
								Item: &ir.Type{
									Kind:      ir.KindPrimitive,
									Primitive: ir.Int,
								},
							},
						},
					},
				},
			},
			expectedFields: map[string]fieldAssertions{
				"NullableArray": {
					fieldType: ir.FieldTypeNullable,
					isArray:   true,
				},
			},
		},
		{
			name: "direct array field",
			irType: &ir.Type{
				Name: "TestType",
				Kind: ir.KindStruct,
				Fields: []*ir.Field{
					{
						Name: "DirectArray",
						Type: &ir.Type{
							Kind: ir.KindArray,
							Item: &ir.Type{
								Kind:      ir.KindPrimitive,
								Primitive: ir.String,
							},
						},
					},
				},
			},
			expectedFields: map[string]fieldAssertions{
				"DirectArray": {
					fieldType: ir.FieldTypeArray,
					isArray:   true,
				},
			},
		},
		{
			name: "optional array of structs",
			irType: &ir.Type{
				Name: "TestType",
				Kind: ir.KindStruct,
				Fields: []*ir.Field{
					{
						Name: "OptionalStructArray",
						Type: &ir.Type{
							Name: "OptStructArray",
							Kind: ir.KindGeneric,
							GenericOf: &ir.Type{
								Kind: ir.KindArray,
								Item: &ir.Type{
									Name: "NestedStruct",
									Kind: ir.KindStruct,
								},
							},
						},
					},
				},
			},
			expectedFields: map[string]fieldAssertions{
				"OptionalStructArray": {
					fieldType:        ir.FieldTypeOptional,
					isArray:          true,
					isArrayOfStructs: true,
				},
			},
		},
		{
			name: "optional map field (not array)",
			irType: &ir.Type{
				Name: "TestType",
				Kind: ir.KindStruct,
				Fields: []*ir.Field{
					{
						Name: "OptionalMap",
						Type: &ir.Type{
							Name: "OptStringMap",
							Kind: ir.KindGeneric,
							GenericOf: &ir.Type{
								Kind: ir.KindMap,
							},
						},
					},
				},
			},
			expectedFields: map[string]fieldAssertions{
				"OptionalMap": {
					fieldType: ir.FieldTypeOptional,
					isMap:     true,
					isArray:   false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Generator{
				equalitySpecs: []*ir.EqualityMethodSpec{},
			}

			spec := g.createEqualityMethodSpec(tt.irType)
			require.NotNil(t, spec, "spec should not be nil")
			require.Equal(t, tt.irType.Name, spec.TypeName)

			for fieldName, assertions := range tt.expectedFields {
				var found *ir.FieldEqualitySpec
				for i := range spec.Fields {
					if spec.Fields[i].FieldName == fieldName {
						found = &spec.Fields[i]
						break
					}
				}

				require.NotNil(t, found, "field %s should exist in spec", fieldName)
				require.Equal(t, assertions.fieldType, found.FieldType,
					"field %s FieldType mismatch", fieldName)
				require.Equal(t, assertions.isArray, found.IsArray,
					"field %s IsArray mismatch", fieldName)
				require.Equal(t, assertions.isArrayOfStructs, found.IsArrayOfStructs,
					"field %s IsArrayOfStructs mismatch", fieldName)
				require.Equal(t, assertions.isMap, found.IsMap,
					"field %s IsMap mismatch", fieldName)
			}
		})
	}
}

type fieldAssertions struct {
	fieldType        ir.FieldTypeCategory
	isArray          bool
	isArrayOfStructs bool
	isMap            bool
}

func TestWriteArrayComparison(t *testing.T) {
	tests := []struct {
		name             string
		aArray           string
		bArray           string
		indent           string
		isArrayOfStructs bool
		hasDepth         bool
		expectedContains []string
	}{
		{
			name:             "primitive array without depth",
			aArray:           "a.Items",
			bArray:           "b.Items",
			indent:           "\t",
			isArrayOfStructs: false,
			hasDepth:         false,
			expectedContains: []string{
				"if len(a.Items) != len(b.Items)",
				"for i := range a.Items",
				"if a.Items[i] != b.Items[i]",
				"return false",
			},
		},
		{
			name:             "struct array with depth",
			aArray:           "a.Nested",
			bArray:           "b.Nested",
			indent:           "\t\t",
			isArrayOfStructs: true,
			hasDepth:         true,
			expectedContains: []string{
				"if len(a.Nested) != len(b.Nested)",
				"for i := range a.Nested",
				"if !a.Nested[i].Equal(b.Nested[i], depth+1)",
				"return false",
			},
		},
		{
			name:             "struct array without depth",
			aArray:           "a.Nested",
			bArray:           "b.Nested",
			indent:           "\t",
			isArrayOfStructs: true,
			hasDepth:         false,
			expectedContains: []string{
				"if len(a.Nested) != len(b.Nested)",
				"for i := range a.Nested",
				"if !a.Nested[i].Equal(b.Nested[i])",
				"return false",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			g := &Generator{}

			g.writeArrayComparison(&b, tt.aArray, tt.bArray, tt.indent, tt.isArrayOfStructs, tt.hasDepth)

			output := b.String()
			for _, expected := range tt.expectedContains {
				require.Contains(t, output, expected,
					"output should contain: %s", expected)
			}

			// Verify proper indentation
			lines := strings.Split(output, "\n")
			for _, line := range lines {
				if line != "" {
					require.True(t, strings.HasPrefix(line, tt.indent),
						"line should start with indent %q: %s", tt.indent, line)
				}
			}
		})
	}
}

func TestWriteFieldComparison_OptionalArray(t *testing.T) {
	tests := []struct {
		name             string
		field            ir.FieldEqualitySpec
		hasDepth         bool
		expectedContains []string
		notContains      []string
	}{
		{
			name: "optional primitive array",
			field: ir.FieldEqualitySpec{
				FieldName:        "Items",
				FieldType:        ir.FieldTypeOptional,
				IsArray:          true,
				IsArrayOfStructs: false,
			},
			hasDepth: false,
			expectedContains: []string{
				"// Compare optional field: Items",
				"if a.Items.Set != b.Items.Set",
				"if a.Items.Set {",
				"if len(a.Items.Value) != len(b.Items.Value)",
				"for i := range a.Items.Value",
				"if a.Items.Value[i] != b.Items.Value[i]",
			},
			notContains: []string{
				"if a.Items.Value != b.Items.Value", // Should NOT use direct slice comparison
			},
		},
		{
			name: "optional struct array with depth",
			field: ir.FieldEqualitySpec{
				FieldName:        "Nested",
				FieldType:        ir.FieldTypeOptional,
				IsArray:          true,
				IsArrayOfStructs: true,
			},
			hasDepth: true,
			expectedContains: []string{
				"// Compare optional field: Nested",
				"if a.Nested.Set != b.Nested.Set",
				"if a.Nested.Set {",
				"if len(a.Nested.Value) != len(b.Nested.Value)",
				"for i := range a.Nested.Value",
				"if !a.Nested.Value[i].Equal(b.Nested.Value[i], depth+1)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			g := &Generator{}

			g.writeFieldComparison(&b, tt.field, tt.hasDepth)

			output := b.String()
			for _, expected := range tt.expectedContains {
				require.Contains(t, output, expected,
					"output should contain: %s", expected)
			}
			for _, notExpected := range tt.notContains {
				require.NotContains(t, output, notExpected,
					"output should NOT contain: %s", notExpected)
			}
		})
	}
}

func TestWriteFieldComparison_NullableArray(t *testing.T) {
	tests := []struct {
		name             string
		field            ir.FieldEqualitySpec
		hasDepth         bool
		expectedContains []string
		notContains      []string
	}{
		{
			name: "nullable primitive array",
			field: ir.FieldEqualitySpec{
				FieldName:        "Items",
				FieldType:        ir.FieldTypeNullable,
				IsArray:          true,
				IsArrayOfStructs: false,
			},
			hasDepth: false,
			expectedContains: []string{
				"// Compare nullable field: Items",
				"if a.Items.Null != b.Items.Null",
				"if !a.Items.Null {",
				"if len(a.Items.Value) != len(b.Items.Value)",
				"for i := range a.Items.Value",
				"if a.Items.Value[i] != b.Items.Value[i]",
			},
			notContains: []string{
				"if a.Items.Value != b.Items.Value", // Should NOT use direct slice comparison
			},
		},
		{
			name: "nullable struct array with depth",
			field: ir.FieldEqualitySpec{
				FieldName:        "Nested",
				FieldType:        ir.FieldTypeNullable,
				IsArray:          true,
				IsArrayOfStructs: true,
			},
			hasDepth: true,
			expectedContains: []string{
				"// Compare nullable field: Nested",
				"if a.Nested.Null != b.Nested.Null",
				"if !a.Nested.Null {",
				"if len(a.Nested.Value) != len(b.Nested.Value)",
				"for i := range a.Nested.Value",
				"if !a.Nested.Value[i].Equal(b.Nested.Value[i], depth+1)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			g := &Generator{}

			g.writeFieldComparison(&b, tt.field, tt.hasDepth)

			output := b.String()
			for _, expected := range tt.expectedContains {
				require.Contains(t, output, expected,
					"output should contain: %s", expected)
			}
			for _, notExpected := range tt.notContains {
				require.NotContains(t, output, notExpected,
					"output should NOT contain: %s", notExpected)
			}
		})
	}
}
