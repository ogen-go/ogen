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
			{Name: "foo", Schema: &Schema{Type: []string{"string"}}},
			{Name: "bar", Schema: &Schema{Type: []string{"number"}}},
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
		{`{"type":"string"}`, AdditionalProperties{Schema: Schema{Type: []string{"string"}}}, false},
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
			{Pattern: "\\w+", Schema: &Schema{Type: []string{"string"}}},
			{Pattern: "\\d+", Schema: &Schema{Type: []string{"number"}}},
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
