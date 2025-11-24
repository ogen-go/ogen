package jsonschema

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCrossTypeConstraints(t *testing.T) {
	tests := []struct {
		name           string
		schema         *RawSchema
		allowCrossType bool
		wantErr        bool
		errContains    string
		checkSchema    func(*testing.T, *Schema)
	}{
		{
			name: "maximum on string type - strict mode",
			schema: &RawSchema{
				Type:    "string",
				Maximum: Num(`1000`),
			},
			allowCrossType: false,
			wantErr:        true,
			errContains:    "unexpected field for type \"string\"",
		},
		{
			name: "maximum on string type - default mode interprets as numeric constraint",
			schema: &RawSchema{
				Type:    "string",
				Maximum: Num(`1000`),
			},
			allowCrossType: true,
			wantErr:        false,
			checkSchema: func(t *testing.T, s *Schema) {
				// Verify the constraint was preserved (will be applied during code generation)
				require.Equal(t, Num(`1000`), s.Maximum)
			},
		},
		{
			name: "minimum on string type - strict mode",
			schema: &RawSchema{
				Type:    "string",
				Minimum: Num(`0.001`),
			},
			allowCrossType: false,
			wantErr:        true,
			errContains:    "unexpected field for type \"string\"",
		},
		{
			name: "minimum on string type - default mode",
			schema: &RawSchema{
				Type:    "string",
				Minimum: Num(`0.001`),
			},
			allowCrossType: true,
			wantErr:        false,
		},
		{
			name: "pattern on number type - strict mode",
			schema: &RawSchema{
				Type:    "number",
				Pattern: `^\d+(\.\d{1,2})?$`,
			},
			allowCrossType: false,
			wantErr:        true,
			errContains:    "unexpected field for type \"number\"",
		},
		{
			name: "pattern on number type - default mode",
			schema: &RawSchema{
				Type:    "number",
				Pattern: `^\d+(\.\d{1,2})?$`,
			},
			allowCrossType: true,
			wantErr:        false,
		},
		{
			name: "pattern on integer type - strict mode",
			schema: &RawSchema{
				Type:    "integer",
				Pattern: `^\d+$`,
			},
			allowCrossType: false,
			wantErr:        true,
			errContains:    "unexpected field for type \"integer\"",
		},
		{
			name: "pattern on integer type - default mode",
			schema: &RawSchema{
				Type:    "integer",
				Pattern: `^\d+$`,
			},
			allowCrossType: true,
			wantErr:        false,
		},
		{
			name: "maxLength on number type - strict mode",
			schema: &RawSchema{
				Type:      "number",
				MaxLength: uint64Ptr(10),
			},
			allowCrossType: false,
			wantErr:        true,
			errContains:    "unexpected field for type \"number\"",
		},
		{
			name: "maxLength on number type - default mode",
			schema: &RawSchema{
				Type:      "number",
				MaxLength: uint64Ptr(10),
			},
			allowCrossType: true,
			wantErr:        false,
		},
		{
			name: "valid schema - both modes",
			schema: &RawSchema{
				Type:      "string",
				MaxLength: uint64Ptr(100),
				Pattern:   `^\w+$`,
			},
			allowCrossType: false,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(Settings{
				AllowCrossTypeConstraints: tt.allowCrossType,
			})

			result, err := parser.Parse(tt.schema, testCtx())
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				if tt.checkSchema != nil {
					tt.checkSchema(t, result)
				}
			}
		})
	}
}

func TestCrossTypeConstraintsComplex(t *testing.T) {
	// Test with a more complex schema like the one in the issue
	schema := &RawSchema{
		Type: "object",
		Properties: []RawProperty{
			{
				Name: "weight",
				Schema: &RawSchema{
					Type:    "string",
					Maximum: Num(`1000`),
					Minimum: Num(`0.001`),
				},
			},
			{
				Name: "quantity",
				Schema: &RawSchema{
					Type:    "number",
					Pattern: `^\d+(\.\d{1,2})?$`,
				},
			},
		},
	}

	t.Run("strict mode fails", func(t *testing.T) {
		parser := NewParser(Settings{
			AllowCrossTypeConstraints: false,
		})
		_, err := parser.Parse(schema, testCtx())
		require.Error(t, err)
		require.Contains(t, err.Error(), "unexpected field for type")
	})

	t.Run("default mode succeeds with interpretation", func(t *testing.T) {
		parser := NewParser(Settings{
			AllowCrossTypeConstraints: true,
		})
		result, err := parser.Parse(schema, testCtx())
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, Object, result.Type)
		require.Len(t, result.Properties, 2)
	})
}

func uint64Ptr(v uint64) *uint64 {
	return &v
}
