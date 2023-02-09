package uri

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCookieParamDecoder(t *testing.T) {
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

			d := &cookieParamDecoder{
				paramName: test.CookieName,
				explode:   test.Explode,
				req:       req,
			}

			s, err := d.DecodeValue()
			require.NoError(t, err, fmt.Sprintf("Test %d", i+1))

			require.Equal(t, test.Expect, s, fmt.Sprintf("Test %d", i+1))
		}
	})

	t.Run("array", func(t *testing.T) {
		tests := []struct {
			CookieName string
			Input      http.Header
			Explode    bool
			Expect     []string
		}{
			{
				CookieName: "token",
				Input: http.Header{
					"Cookie": []string{"token=3%2C4%2C5"},
				},
				Explode: false,
				Expect:  []string{"3", "4", "5"},
			},
		}
		for i, test := range tests {
			req := &http.Request{
				Header: test.Input,
			}

			d := &cookieParamDecoder{
				paramName: test.CookieName,
				explode:   test.Explode,
				req:       req,
			}

			var items []string
			err := d.DecodeArray(func(d Decoder) error {
				v, err := d.DecodeValue()
				if err != nil {
					return err
				}

				items = append(items, v)
				return nil
			})
			require.NoError(t, err, fmt.Sprintf("Test %d", i+1))

			require.Equal(t, test.Expect, items, fmt.Sprintf("Test %d", i+1))
		}
	})

	t.Run("object", func(t *testing.T) {
		tests := []struct {
			CookieName string
			Input      http.Header
			Explode    bool
			Expect     []Field
		}{
			{
				CookieName: "token",
				Input: http.Header{
					"Cookie": []string{"token=role%2Cadmin%2CfirstName%2CAlex"},
				},
				Explode: false,
				Expect: []Field{
					{
						Name:  "role",
						Value: "admin",
					},
					{
						Name:  "firstName",
						Value: "Alex",
					},
				},
			},
		}
		for i, test := range tests {
			req := &http.Request{
				Header: test.Input,
			}

			d := &cookieParamDecoder{
				paramName: test.CookieName,
				explode:   test.Explode,
				req:       req,
			}

			var fields []Field
			err := d.DecodeFields(func(field string, d Decoder) error {
				v, err := d.DecodeValue()
				if err != nil {
					return err
				}

				fields = append(fields, Field{
					Name:  field,
					Value: v,
				})
				return nil
			})
			require.NoError(t, err, fmt.Sprintf("Test %d", i+1))

			require.Equal(t, test.Expect, fields, fmt.Sprintf("Test %d", i+1))
		}
	})
}
