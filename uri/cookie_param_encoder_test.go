package uri

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCookieParamEncoder(t *testing.T) {
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

			e := &cookieParamEncoder{
				receiver:  newReceiver(),
				paramName: test.CookieName,
				explode:   test.Explode,
				req:       req,
			}

			require.NoError(t, e.EncodeValue(test.Input))
			e.serialize()

			require.Equal(t, test.Expect, req.Header, fmt.Sprintf("Test %d", i+1))
		}
	})
	t.Run("array", func(t *testing.T) {
		tests := []struct {
			CookieName string
			Input      []string
			Explode    bool
			Expect     http.Header
		}{
			{
				CookieName: "token",
				Input:      []string{"3", "4", "5"},
				Explode:    false,
				Expect: http.Header{
					"Cookie": []string{"token=3%2C4%2C5"},
				},
			},
		}
		for i, test := range tests {
			req := &http.Request{
				Header: http.Header{},
			}

			e := &cookieParamEncoder{
				receiver:  newReceiver(),
				paramName: test.CookieName,
				explode:   test.Explode,
				req:       req,
			}

			require.NoError(t, e.EncodeArray(func(e Encoder) error {
				for _, v := range test.Input {
					if err := e.EncodeValue(v); err != nil {
						return err
					}
				}
				return nil
			}))
			e.serialize()

			require.Equal(t, test.Expect, req.Header, fmt.Sprintf("Test %d", i+1))
		}
	})

	t.Run("object", func(t *testing.T) {
		tests := []struct {
			CookieName string
			Input      []Field
			Explode    bool
			Expect     http.Header
		}{
			{
				CookieName: "token",
				Input: []Field{
					{
						Name:  "role",
						Value: "admin",
					},
					{
						Name:  "firstName",
						Value: "Alex",
					},
				},
				Explode: false,
				Expect: http.Header{
					"Cookie": []string{"token=role%2Cadmin%2CfirstName%2CAlex"},
				},
			},
		}
		for i, test := range tests {
			req := &http.Request{
				Header: http.Header{},
			}

			e := &cookieParamEncoder{
				receiver:  newReceiver(),
				paramName: test.CookieName,
				explode:   test.Explode,
				req:       req,
			}

			for _, f := range test.Input {
				require.NoError(t, e.EncodeField(f.Name, func(e Encoder) error {
					return e.EncodeValue(f.Value)
				}))
			}
			e.serialize()

			require.Equal(t, test.Expect, req.Header, fmt.Sprintf("Test %d", i+1))
		}
	})
}
