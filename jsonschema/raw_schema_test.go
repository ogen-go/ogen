package jsonschema

import (
	"fmt"
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNum_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		data    string
		value   Num
		wantErr bool
	}{
		{`0`, Num(`0`), false},
		{`1e1`, Num(`10`), false},
		{`0x0a`, Num(`10`), false},
		// Invalid YAML.
		{`"`, nil, true},
		{`0ee1`, nil, true},
		// Invalid type.
		{`{}`, nil, true},
		{`"100"`, nil, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			var val Num
			err := yaml.Unmarshal([]byte(tt.data), &val)
			if tt.wantErr {
				a.Error(err)
				t.Log("Input:", tt.data)
				t.Log("Error:", err)
				return
			}
			a.NoError(err)
			a.Equal(tt.value, val)
		})
	}
}

func TestNum_UnmarshalNextJSON(t *testing.T) {
	tests := []struct {
		data    string
		value   Num
		wantErr bool
	}{
		{`0`, Num(`0`), false},
		{`1e1`, Num(`1e1`), false},
		// Invalid JSON.
		{``, nil, true},
		{`"`, nil, true},
		{`0ee1`, nil, true},
		// Invalid type.
		{`{}`, nil, true},
		{`"100"`, nil, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			var val Num
			err := json.Unmarshal([]byte(tt.data), &val)
			if tt.wantErr {
				a.Error(err)
				t.Log("Input:", tt.data)
				t.Log("Error:", err)
				return
			}
			a.NoError(err)
			a.Equal(tt.value, val)

			data, err := json.Marshal(val)
			a.NoError(err)
			a.JSONEq(tt.data, string(data))
		})
	}
}

func TestEnum_UnmarshalNextJSON(t *testing.T) {
	tests := []struct {
		data    string
		value   Enum
		wantErr bool
	}{
		{`[{}, {"a":"b"}]`, Enum{
			json.RawValue(`{}`),
			json.RawValue(`{"a":"b"}`),
		}, false},
		{`[]`, Enum{}, false},
		// Invalid JSON.
		{``, nil, true},
		{`[`, nil, true},
		{`[}`, nil, true},
		// Invalid type.
		{`{}`, nil, true},
		// Duplicate values.
		{`[{}, {}]`, nil, true},
		{`[1.0, 1.0]`, nil, true},
		{`[null, null]`, nil, true},
		{`[{"b":"a","a":"b"}, {"a":"b", "b":"a"}]`, nil, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			var val Enum
			err := json.Unmarshal([]byte(tt.data), &val)
			if tt.wantErr {
				a.Error(err)
				t.Log("Input:", tt.data)
				t.Log("Error:", err)
				return
			}
			a.NoError(err)
			a.Equal(tt.value, val)

			data, err := json.Marshal(val)
			a.NoError(err)
			a.JSONEq(tt.data, string(data))
		})
	}
}

func TestRawProperties_UnmarshalNextJSON(t *testing.T) {
	tests := []struct {
		data    string
		value   RawProperties
		wantErr bool
	}{

		{`{"foo":{"type":"string"}, "bar":{"type":"number"}}`, RawProperties{
			{Name: "foo", Schema: &RawSchema{Type: "string"}},
			{Name: "bar", Schema: &RawSchema{Type: "number"}},
		}, false},
		// Invalid JSON.
		{`{0:"string"}`, RawProperties{}, true},
		{``, RawProperties{}, true},
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
			err := json.Unmarshal([]byte(tt.data), &val)
			if tt.wantErr {
				a.Error(err)
				t.Log("Input:", tt.data)
				t.Log("Error:", err)
				return
			}
			a.NoError(err)
			a.Equal(tt.value, val)

			data, err := json.Marshal(val)
			a.NoError(err)
			a.JSONEq(tt.data, string(data))
		})
	}
}

func TestAdditionalProperties_UnmarshalNextJSON(t *testing.T) {
	tests := []struct {
		data    string
		value   AdditionalProperties
		wantErr bool
	}{
		{`{"type":"string"}`, AdditionalProperties{Schema: RawSchema{Type: "string"}}, false},
		{`false`, AdditionalProperties{Bool: new(bool)}, false},
		// Invalid JSON
		{`{0:"string"}`, AdditionalProperties{}, true},
		{``, AdditionalProperties{}, true},
		{`{`, AdditionalProperties{}, true},
		{`{]`, AdditionalProperties{}, true},
		// Invalid type.
		{`[]`, AdditionalProperties{}, true},
		{`{"type":10}`, AdditionalProperties{}, true},
		{`0`, AdditionalProperties{}, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			var val AdditionalProperties
			err := json.Unmarshal([]byte(tt.data), &val)
			if tt.wantErr {
				a.Error(err)
				t.Log("Input:", tt.data)
				t.Log("Error:", err)
				return
			}
			a.NoError(err)
			a.Equal(tt.value, val)

			data, err := json.Marshal(val)
			a.NoError(err)
			a.JSONEq(tt.data, string(data))
		})
	}
}

func TestRawPatternProperties_UnmarshalNextJSON(t *testing.T) {
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
		{`{0:"string"}`, RawPatternProperties{}, true},
		{``, RawPatternProperties{}, true},
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
			err := json.Unmarshal([]byte(tt.data), &val)
			if tt.wantErr {
				a.Error(err)
				t.Log("Input:", tt.data)
				t.Log("Error:", err)
				return
			}
			a.NoError(err)
			a.Equal(tt.value, val)

			data, err := json.Marshal(val)
			a.NoError(err)
			a.JSONEq(tt.data, string(data))
		})
	}
}
