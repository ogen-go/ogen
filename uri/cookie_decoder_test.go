package uri

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCookieDecoder(t *testing.T) {
	t.Run("primitive", func(t *testing.T) {
		tests := []struct {
			CookieName string
			Input      http.Header
			Explode    bool
			Expect     string
		}{
			{
				CookieName: "token",
				Input: http.Header{
					"Cookie": []string{"token=foobar"},
				},
				Explode: false,
				Expect:  "foobar",
			},
			{
				CookieName: "token",
				Input: http.Header{
					"Cookie": []string{"token=foobar"},
				},
				Explode: true,
				Expect:  "foobar",
			},

			// Escaping
			{
				CookieName: "token",
				Input: http.Header{
					"Cookie": []string{"token=`%22'%3B%2C./<>?[]{}%5C|~!@#$%25^&*()_+-="},
				},
				Explode: false,
				Expect:  "`\"';,./<>?[]{}\\|~!@#$%^&*()_+-=",
			},
			{
				CookieName: "token",
				Input: http.Header{
					"Cookie": []string{"token=`%22'%3B%2C./<>?[]{}%5C|~!@#$%25^&*()_+-="},
				},
				Explode: true,
				Expect:  "`\"';,./<>?[]{}\\|~!@#$%^&*()_+-=",
			},
		}
		for i, test := range tests {
			req := &http.Request{
				Header: test.Input,
			}

			dec := NewCookieDecoder(req)
			cfg := CookieParameterDecodingConfig{
				Name:    test.CookieName,
				Explode: test.Explode,
			}

			var s string
			err := dec.DecodeParam(cfg, func(dec Decoder) (err error) {
				s, err = dec.DecodeValue()
				return err
			})
			require.NoError(t, err, fmt.Sprintf("Test %d", i+1))
			require.Equal(t, test.Expect, s, fmt.Sprintf("Test %d", i+1))
		}
	})
}
