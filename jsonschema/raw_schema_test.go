package jsonschema

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnum_UnmarshalJSON(t *testing.T) {
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
			if err := e.UnmarshalJSON([]byte(tt.data)); tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
