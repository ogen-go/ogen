package jsonschema

import (
	"fmt"
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/stretchr/testify/require"
)

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
