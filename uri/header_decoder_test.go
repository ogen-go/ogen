package uri

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHeaderDecoder_HasParam(t *testing.T) {
	tests := []struct {
		Input      http.Header
		HeaderName string
		WantErr    string
	}{
		{
			Input:      http.Header{},
			HeaderName: "X-Foo",
			WantErr:    "header parameter \"X-Foo\" not set",
		},
		{
			Input: http.Header{
				"X-Foo": []string{"bar"},
			},
			HeaderName: "X-Foo",
		},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			d := NewHeaderDecoder(tt.Input)

			err := d.HasParam(HeaderParameterDecodingConfig{
				Name: tt.HeaderName,
			})
			if tt.WantErr != "" {
				require.EqualError(t, err, tt.WantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
