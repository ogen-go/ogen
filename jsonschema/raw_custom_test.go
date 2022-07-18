package jsonschema

import (
	"encoding/json"
	"fmt"
	"testing"

	yaml "github.com/go-faster/yamlx"
	"github.com/stretchr/testify/require"
)

type encoding struct {
	marshal   func(interface{}) ([]byte, error)
	unmarshal func([]byte, interface{}) error
	compare   func(a *require.Assertions, got, want string, msgArgs ...interface{})
}

func testCustomEncoding(
	createVal func() interface{},
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
	createVal func() interface{},
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

func TestRawProperties(t *testing.T) {
	create := func() interface{} {
		return &RawProperties{}
	}

	tests := []struct {
		data    string
		value   RawProperties
		wantErr bool
	}{
		{`{"foo":{"type":"string"}, "bar":{"type":"number"}}`, RawProperties{
			{Name: "foo", Schema: &RawSchema{Type: "string"}},
			{Name: "bar", Schema: &RawSchema{Type: "number"}},
		}, false},
		// Invalid YAML.
		{`{`, RawProperties{}, true},
		{`{]`, RawProperties{}, true},
		// Invalid type.
		{`{"foobar":"string"}`, RawProperties{}, true},
		{`0`, RawProperties{}, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), testCustomEncodings(create, tt.data, tt.wantErr))
	}
}

func TestAdditionalProperties(t *testing.T) {
	create := func() interface{} {
		return &AdditionalProperties{}
	}

	tests := []struct {
		data    string
		value   AdditionalProperties
		wantErr bool
	}{
		{`{"type":"string"}`, AdditionalProperties{Schema: RawSchema{Type: "string"}}, false},
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
	create := func() interface{} {
		return &RawPatternProperties{}
	}

	tests := []struct {
		data    string
		value   RawPatternProperties
		wantErr bool
	}{
		{`{"\\w+":{"type":"string"}, "\\d+":{"type":"number"}}`, RawPatternProperties{
			{Pattern: "\\w+", Schema: &RawSchema{Type: "string"}},
			{Pattern: "\\d+", Schema: &RawSchema{Type: "number"}},
		}, false},
		// Invalid JSON.
		{`{`, RawPatternProperties{}, true},
		{`{]`, RawPatternProperties{}, true},
		// Invalid type.
		{`{"^[a-zA-Z0-9]*$":"string"}`, RawPatternProperties{}, true},
		{`0`, RawPatternProperties{}, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), testCustomEncodings(create, tt.data, tt.wantErr))
	}
}
