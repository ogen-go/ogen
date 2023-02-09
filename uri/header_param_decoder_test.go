package uri

import (
	"bufio"
	"fmt"
	"net/http"
	"net/textproto"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHeaderParamDecoder(t *testing.T) {
	t.Run("primitive", func(t *testing.T) {
		tests := []struct {
			HeaderName string
			Input      string
			Explode    bool
			Expect     string
		}{
			{
				HeaderName: "X-MyHeader",
				Input:      "X-MyHeader: 5\r\n\r\n",
				Explode:    false,
				Expect:     "5",
			},
			{
				HeaderName: "X-MyHeader",
				Input:      "X-MyHeader: 5\r\n\r\n",
				Explode:    true,
				Expect:     "5",
			},
		}
		for i, test := range tests {
			h, err := textproto.NewReader(bufio.NewReader(strings.NewReader(test.Input))).ReadMIMEHeader()
			require.NoError(t, err, fmt.Sprintf("Test %d", i+1))

			result, err := (&headerParamDecoder{
				paramName: test.HeaderName,
				explode:   test.Explode,
				header:    http.Header(h),
			}).DecodeValue()

			require.NoError(t, err, fmt.Sprintf("Test %d", i+1))

			require.Equal(t, test.Expect, result, fmt.Sprintf("Test %d", i+1))
		}
	})

	t.Run("array", func(t *testing.T) {
		tests := []struct {
			HeaderName string
			Input      string
			Explode    bool
			Expect     []string
		}{
			{
				HeaderName: "X-MyHeader",
				Input:      "X-MyHeader: 3,4,5\r\n\r\n",
				Explode:    false,
				Expect:     []string{"3", "4", "5"},
			},
			{
				HeaderName: "X-MyHeader",
				Input:      "X-MyHeader: 3,4,5\r\n\r\n",
				Explode:    true,
				Expect:     []string{"3", "4", "5"},
			},
		}
		for i, test := range tests {
			h, err := textproto.NewReader(bufio.NewReader(strings.NewReader(test.Input))).ReadMIMEHeader()
			require.NoError(t, err, fmt.Sprintf("Test %d", i+1))

			var items []string
			err = (&headerParamDecoder{
				paramName: test.HeaderName,
				explode:   test.Explode,
				header:    http.Header(h),
			}).DecodeArray(func(d Decoder) error {
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
			HeaderName string
			Input      string
			Explode    bool
			Expect     []Field
		}{
			{
				HeaderName: "X-MyHeader",
				Input:      "X-MyHeader: role,admin,firstName,Alex\r\n\r\n",
				Explode:    false,
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
			{
				HeaderName: "X-MyHeader",
				Input:      "X-MyHeader: role=admin,firstName=Alex\r\n\r\n",
				Explode:    true,
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
			h, err := textproto.NewReader(bufio.NewReader(strings.NewReader(test.Input))).ReadMIMEHeader()
			require.NoError(t, err, fmt.Sprintf("Test %d", i+1))

			var fields []Field
			err = (&headerParamDecoder{
				paramName: test.HeaderName,
				explode:   test.Explode,
				header:    http.Header(h),
			}).DecodeFields(func(field string, d Decoder) error {
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
