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
		wantErr bool
	}{
		{`[{}, {}]`, true},
		{`[1.0, 1.0]`, true},
		{`[null, null]`, true},
		{`[{}, {"a":"b"}]`, false},
		{`[{"b":"a","a":"b"}, {"a":"b", "b":"a"}]`, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			var e Enum
			if err := json.Unmarshal([]byte(tt.data), &e); tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
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
				return
			}
			a.NoError(err)
			a.Equal(tt.value, val)

			var s RawSchema
			err = json.Unmarshal([]byte(fmt.Sprintf(`{"additionalProperties":%s}`, tt.data)), &s)
			a.NoError(err)
			a.Equal(&tt.value, s.AdditionalProperties)
		})
	}
}
