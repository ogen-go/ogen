package ogen

import (
	"fmt"
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/stretchr/testify/require"
)

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

func TestProperties_UnmarshalNextJSON(t *testing.T) {
	tests := []struct {
		data    string
		value   Properties
		wantErr bool
	}{

		{`{"foo":{"type":"string"}, "bar":{"type":"number"}}`, Properties{
			{Name: "foo", Schema: &Schema{Type: "string"}},
			{Name: "bar", Schema: &Schema{Type: "number"}},
		}, false},
		// Invalid JSON.
		{`{0:"string"}`, Properties{}, true},
		{``, Properties{}, true},
		{`{`, Properties{}, true},
		{`{]`, Properties{}, true},
		// Invalid type.
		{`{"foobar":"string"}`, Properties{}, true},
		{`0`, Properties{}, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			var val Properties
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
		{`{"type":"string"}`, AdditionalProperties{Schema: Schema{Type: "string"}}, false},
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

func TestPatternProperties_UnmarshalNextJSON(t *testing.T) {
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
		{`{0:"string"}`, PatternProperties{}, true},
		{``, PatternProperties{}, true},
		{`{`, PatternProperties{}, true},
		{`{]`, PatternProperties{}, true},
		// Invalid type.
		{`{"^[a-zA-Z0-9]*$":"string"}`, PatternProperties{}, true},
		{`0`, PatternProperties{}, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			var val PatternProperties
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
