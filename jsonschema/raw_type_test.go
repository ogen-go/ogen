package jsonschema

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-faster/yaml"
	"github.com/stretchr/testify/require"
)

func TestRawSchemaTypeList(t *testing.T) {
	tests := []struct {
		input        string
		wantType     string
		wantNullable bool
		wantErr      bool
	}{
		{`{"type": "string"}`, "string", false, false},
		{`{"type": "string", "nullable": true}`, "string", true, false},
		{`{"type": ["string", "null"]}`, "string", true, false},
		{`{"type": ["null", "integer"]}`, "integer", true, false},
		{`{"type": ["string"]}`, "string", false, false},
		{`{"type": ["null"]}`, "null", false, false},
		{`{"type": ["string", "string", "null"]}`, "string", true, false},
		{`{"description": "no type"}`, "", false, false},
		// Multiple non-null types are not representable.
		{`{"type": ["string", "integer"]}`, "", false, true},
		{`{"type": []}`, "", false, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			t.Run("YAML", func(t *testing.T) {
				a := require.New(t)

				s := RawSchema{}
				err := yaml.Unmarshal([]byte(tt.input), &s)
				if tt.wantErr {
					a.Error(err)
					return
				}
				a.NoError(err)
				a.Equal(tt.wantType, s.Type)
				a.Equal(tt.wantNullable, s.Nullable)
			})
			t.Run("JSON", func(t *testing.T) {
				a := require.New(t)

				s := RawSchema{}
				err := json.Unmarshal([]byte(tt.input), &s)
				if tt.wantErr {
					a.Error(err)
					return
				}
				a.NoError(err)
				a.Equal(tt.wantType, s.Type)
				a.Equal(tt.wantNullable, s.Nullable)
			})
		})
	}
}

func TestRawSchemaTypeListYAML(t *testing.T) {
	tests := []struct {
		input        string
		wantType     string
		wantNullable bool
	}{
		// Unquoted null in a type list.
		{"type: [string, null]", "string", true},
		// Block style list.
		{"type:\n  - string\n  - 'null'\ndescription: foo", "string", true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			s := RawSchema{}
			a.NoError(yaml.Unmarshal([]byte(tt.input), &s))
			a.Equal(tt.wantType, s.Type)
			a.Equal(tt.wantNullable, s.Nullable)
		})
	}
}

func TestRawSchemaTypeListNested(t *testing.T) {
	a := require.New(t)

	const input = `
type: [object, 'null']
properties:
  name:
    type: [string, 'null']
  id:
    type: integer
`
	s := RawSchema{}
	a.NoError(yaml.Unmarshal([]byte(input), &s))
	a.Equal("object", s.Type)
	a.True(s.Nullable)

	a.Len(s.Properties, 2)
	a.Equal("string", s.Properties[0].Schema.Type)
	a.True(s.Properties[0].Schema.Nullable)
	a.Equal("integer", s.Properties[1].Schema.Type)
	a.False(s.Properties[1].Schema.Nullable)
}

// Collapsing must not modify the original document: reference resolvers
// decode the same nodes again.
func TestCollapseTypeNodeKeepsOriginal(t *testing.T) {
	a := require.New(t)

	const input = `type: [string, 'null']`
	var root yaml.Node
	a.NoError(yaml.Unmarshal([]byte(input), &root))

	for range 2 {
		s := RawSchema{}
		a.NoError(root.Decode(&s))
		a.Equal("string", s.Type)
		a.True(s.Nullable)
	}
}
