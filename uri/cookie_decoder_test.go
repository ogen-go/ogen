package uri

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCookieDecoder_HasParam(t *testing.T) {
	tests := []struct {
		Input      http.Header
		CookieName string
		WantErr    string
	}{
		{
			Input:      http.Header{},
			CookieName: "foo",
			WantErr:    "invalid: foo (field required)",
		},
		{
			Input: http.Header{
				"Cookie": []string{"foo=bar"},
			},
			CookieName: "foo",
		},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			req := &http.Request{
				Header: tt.Input,
			}
			d := NewCookieDecoder(req)

			err := d.HasParam(CookieParameterDecodingConfig{
				Name: tt.CookieName,
			})
			if tt.WantErr != "" {
				require.EqualError(t, err, tt.WantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
