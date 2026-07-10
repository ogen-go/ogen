package ogen

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-faster/yaml"
	"github.com/stretchr/testify/require"
)

type encoding struct {
	marshal   func(any) ([]byte, error)
	unmarshal func([]byte, any) error
	compare   func(a *require.Assertions, got, want string, msgArgs ...any)
}

func testCustomEncoding(
	createVal func() any,
	input string,
	wantErr bool,
	e encoding,
) func(t *testing.T) {
	return func(t *testing.T) {
		a := require.New(t)

		val := createVal()
		err := e.unmarshal([]byte(input), val)
		if wantErr {
			a.Error(err)
			t.Logf("Input: %q", input)
			t.Logf("Error: %+v", err)
			return
		}
		a.NoError(err)

		data, err := e.marshal(val)
		a.NoError(err)
		e.compare(a, input, string(data))
	}
}

func testCustomEncodings(
	createVal func() any,
	input string,
	wantErr bool,
) func(t *testing.T) {
	js := encoding{
		marshal:   json.Marshal,
		unmarshal: json.Unmarshal,
		compare:   (*require.Assertions).JSONEq,
	}
	yml := encoding{
		marshal:   yaml.Marshal,
		unmarshal: yaml.Unmarshal,
		compare:   (*require.Assertions).YAMLEq,
	}

	return func(t *testing.T) {
		t.Run("YAML", testCustomEncoding(
			createVal,
			input,
			wantErr,
			yml,
		))
		t.Run("JSON", testCustomEncoding(
			createVal,
			input,
			wantErr,
			js,
		))
	}
}

func TestProperties(t *testing.T) {
	create := func() any {
		return &Properties{}
	}

	tests := []struct {
		data    string
		value   Properties
		wantErr bool
	}{
		{`{"foo":{"type":"string"}, "bar":{"type":"number"}}`, Properties{
			{Name: "foo", Schema: &Schema{Type: "string"}},
			{Name: "bar", Schema: &Schema{Type: "number"}},
		}, false},
		// Invalid YAML.
		{`{`, Properties{}, true},
		{`{]`, Properties{}, true},
		// Invalid type.
		{`{"foobar":"string"}`, Properties{}, true},
		{`0`, Properties{}, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), testCustomEncodings(create, tt.data, tt.wantErr))
	}
}

func TestAdditionalProperties(t *testing.T) {
	create := func() any {
		return &AdditionalProperties{}
	}

	tests := []struct {
		data    string
		value   AdditionalProperties
		wantErr bool
	}{
		{`{"type":"string"}`, AdditionalProperties{Schema: Schema{Type: "string"}}, false},
		{`false`, AdditionalProperties{Bool: new(bool)}, false},
		// Invalid YAML.
		{`{`, AdditionalProperties{}, true},
		{`{]`, AdditionalProperties{}, true},
		// Invalid type.
		{`[]`, AdditionalProperties{}, true},
		{`{"type": {}}`, AdditionalProperties{}, true},
		{`0`, AdditionalProperties{}, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), testCustomEncodings(create, tt.data, tt.wantErr))
	}
}

func TestPatternProperties(t *testing.T) {
	create := func() any {
		return &PatternProperties{}
	}

	tests := []struct {
		data    string
		value   PatternProperties
		wantErr bool
	}{
		{`{"\\w+":{"type":"string"}, "\\d+":{"type":"number"}}`, PatternProperties{
			{Pattern: "\\w+", Schema: &Schema{Type: "string"}},
			{Pattern: "\\d+", Schema: &Schema{Type: "number"}},
		}, false},
		// Invalid JSON.
		{`{`, PatternProperties{}, true},
		{`{]`, PatternProperties{}, true},
		// Invalid type.
		{`{"^[a-zA-Z0-9]*$":"string"}`, PatternProperties{}, true},
		{`0`, PatternProperties{}, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), testCustomEncodings(create, tt.data, tt.wantErr))
	}
}

func TestSchemaTypeList(t *testing.T) {
	t.Run("YAML", func(t *testing.T) {
		a := require.New(t)

		s := Schema{}
		a.NoError(yaml.Unmarshal([]byte(`type: [string, 'null']`), &s))
		a.Equal("string", s.Type)
		a.True(s.Nullable)

		a.Error(yaml.Unmarshal([]byte(`type: [string, integer]`), &Schema{}))
	})
	t.Run("JSON", func(t *testing.T) {
		a := require.New(t)

		s := Schema{}
		a.NoError(json.Unmarshal([]byte(`{"type": ["string", "null"]}`), &s))
		a.Equal("string", s.Type)
		a.True(s.Nullable)

		a.Error(json.Unmarshal([]byte(`{"type": ["string", "integer"]}`), &Schema{}))
	})
}

func TestSpecTypeList(t *testing.T) {
	a := require.New(t)

	const input = `
openapi: 3.1.0
info:
  title: API
  version: 0.1.0
paths: {}
components:
  schemas:
    Template:
      type: object
      properties:
        archived_at:
          type:
            - string
            - 'null'
          description: Date and time when the template was archived.
        created_at:
          type: string
`
	spec, err := Parse([]byte(input))
	a.NoError(err)

	props := spec.Components.Schemas["Template"].Properties
	a.Equal("archived_at", props[0].Name)
	a.Equal("string", props[0].Schema.Type)
	a.True(props[0].Schema.Nullable)
	a.Equal("created_at", props[1].Name)
	a.Equal("string", props[1].Schema.Type)
	a.False(props[1].Schema.Nullable)
}
