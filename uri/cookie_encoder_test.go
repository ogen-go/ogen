package uri

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCookieEncoder(t *testing.T) {
	t.Run("primitive", func(t *testing.T) {
		tests := []struct {
			CookieName string
			Input      string
			Explode    bool
			Expect     http.Header
		}{
			{
				CookieName: "token",
				Input:      "foobar",
				Explode:    false,
				Expect: http.Header{
					"Cookie": []string{"token=foobar"},
				},
			},
			{
				CookieName: "token",
				Input:      "foobar",
				Explode:    true,
				Expect: http.Header{
					"Cookie": []string{"token=foobar"},
				},
			},

			// Escaping
			{
				CookieName: "token",
				Input:      "`\"';,./<>?[]{}\\|~!@#$%^&*()_+-=",
				Explode:    false,
				Expect: http.Header{
					"Cookie": []string{"token=`%22'%3B%2C./<>?[]{}%5C|~!@#$%25^&*()_+-="},
				},
			},
			{
				CookieName: "token",
				Input:      "`\"';,./<>?[]{}\\|~!@#$%^&*()_+-=",
				Explode:    true,
				Expect: http.Header{
					"Cookie": []string{"token=`%22'%3B%2C./<>?[]{}%5C|~!@#$%25^&*()_+-="},
				},
			},
		}
		for i, test := range tests {
			req := &http.Request{
				Header: http.Header{},
			}

			e := NewCookieEncoder(req)
			cfg := CookieParameterEncodingConfig{
				Name:    test.CookieName,
				Explode: test.Explode,
			}

			require.NoError(t, e.EncodeParam(cfg, func(e Encoder) error {
				return e.EncodeValue(test.Input)
			}))
			require.Equal(t, test.Expect, req.Header, fmt.Sprintf("Test %d", i+1))
		}
	})
}
