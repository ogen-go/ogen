package jsonschema

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestRawProperties(t *testing.T) {
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
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			var val RawProperties
			err := yaml.Unmarshal([]byte(tt.data), &val)
			if tt.wantErr {
				a.Error(err)
				t.Log("Input:", tt.data)
				t.Log("Error:", err)
				return
			}
			a.NoError(err)

			data, err := json.Marshal(val)
			a.NoError(err)
			a.JSONEq(tt.data, string(data))
		})
	}
}

func TestAdditionalProperties(t *testing.T) {
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
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			var val AdditionalProperties
			err := yaml.Unmarshal([]byte(tt.data), &val)
			if tt.wantErr {
				a.Error(err)
				t.Log("Input:", tt.data)
				t.Log("Error:", err)
				return
			}
			a.NoError(err)

			data, err := json.Marshal(val)
			a.NoError(err)
			a.JSONEq(tt.data, string(data))
		})
	}
}

func TestRawPatternProperties(t *testing.T) {
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
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			var val RawPatternProperties
			err := yaml.Unmarshal([]byte(tt.data), &val)
			if tt.wantErr {
				a.Error(err)
				t.Log("Input:", tt.data)
				t.Log("Error:", err)
				return
			}
			a.NoError(err)

			data, err := json.Marshal(val)
			a.NoError(err)
			a.JSONEq(tt.data, string(data))
		})
	}
}
