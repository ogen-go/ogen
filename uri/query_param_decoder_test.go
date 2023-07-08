package uri

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueryParamDecoder(t *testing.T) {
	t.Run("Value", func(t *testing.T) {
		tests := []struct {
			Param   string
			Input   string
			Expect  string
			Style   QueryStyle
			Explode bool
		}{
			{
				Param:   "id",
				Input:   "id=3",
				Expect:  "3",
				Style:   QueryStyleForm,
				Explode: true,
			},
			{
				Param:   "id",
				Input:   "id=3",
				Expect:  "3",
				Style:   QueryStyleForm,
				Explode: false,
			},
		}

		for i, test := range tests {
			values, err := url.ParseQuery(test.Input)
			require.NoError(t, err)

			result, err := (&queryParamDecoder{
				values:    values,
				paramName: test.Param,
				style:     test.Style,
				explode:   test.Explode,
			}).DecodeValue()
			require.NoError(t, err, fmt.Sprintf("Test %d", i+1))
			require.Equal(t, test.Expect, result, fmt.Sprintf("Test %d", i+1))
		}
	})

	t.Run("Array", func(t *testing.T) {
		tests := []struct {
			Param   string
			Input   string
			Expect  []string
			Style   QueryStyle
			Explode bool
		}{
			{
				Param:   "id",
				Input:   "id=3&id=4&id=5",
				Expect:  []string{"3", "4", "5"},
				Style:   QueryStyleForm,
				Explode: true,
			},
			{
				Param:   "id",
				Input:   "id=3%2C4%2C5",
				Expect:  []string{"3", "4", "5"},
				Style:   QueryStyleForm,
				Explode: false,
			},
			{
				Param:   "id",
				Input:   "id=",
				Expect:  []string{""},
				Style:   QueryStyleForm,
				Explode: true,
			},
			{
				Param:   "id",
				Input:   "id=",
				Expect:  []string(nil),
				Style:   QueryStyleForm,
				Explode: false,
			},
			{
				Param:   "id",
				Input:   "id=3&id=4&id=5",
				Expect:  []string{"3", "4", "5"},
				Style:   QueryStyleSpaceDelimited,
				Explode: true,
			},
			// {
			// 	Input:   []string{"a%20b%20c"},
			// 	Expect:  []string{"a", "b", "c"},
			// 	Style:   QueryStyleSpaceDelimited,
			// 	Explode: false,
			// },
			{
				Param:   "id",
				Input:   "id=3&id=4&id=5",
				Expect:  []string{"3", "4", "5"},
				Style:   QueryStylePipeDelimited,
				Explode: true,
			},
			{
				Param:   "id",
				Input:   "id=3%7C4%7C5",
				Expect:  []string{"3", "4", "5"},
				Style:   QueryStylePipeDelimited,
				Explode: false,
			},
		}

		for i, test := range tests {
			values, err := url.ParseQuery(test.Input)
			require.NoError(t, err)
			d := &queryParamDecoder{
				values:    values,
				paramName: test.Param,
				style:     test.Style,
				explode:   test.Explode,
			}

			var items []string
			err = d.DecodeArray(func(d Decoder) error {
				item, err := d.DecodeValue()
				if err != nil {
					return err
				}
				items = append(items, item)
				return nil
			})
			require.NoError(t, err, fmt.Sprintf("Test %d", i+1))
			require.Equal(t, test.Expect, items, fmt.Sprintf("Test %d", i+1))
		}
	})

	t.Run("Object", func(t *testing.T) {
		tests := []struct {
			Param   string
			Input   string
			Expect  []Field
			Style   QueryStyle
			Explode bool
		}{
			{
				Param:   "id",
				Input:   "firstName=Alex&role=admin",
				Style:   QueryStyleForm,
				Explode: true,
				Expect: []Field{
					{"role", "admin"},
					{"firstName", "Alex"},
				},
			},
			{
				Param:   "id",
				Input:   "id=role%2Cadmin%2CfirstName%2CAlex",
				Style:   QueryStyleForm,
				Explode: false,
				Expect: []Field{
					{"role", "admin"},
					{"firstName", "Alex"},
				},
			},
			{
				Param:   "id",
				Input:   "id%5BfirstName%5D=Alex&id%5Brole%5D=admin",
				Style:   QueryStyleDeepObject,
				Explode: true,
				Expect: []Field{
					{"role", "admin"},
					{"firstName", "Alex"},
				},
			},
		}

		for i, test := range tests {
			values, err := url.ParseQuery(test.Input)
			require.NoError(t, err)

			var fields []QueryParameterObjectField
			for _, f := range test.Expect {
				fields = append(fields, QueryParameterObjectField{
					Name:     f.Name,
					Required: true,
				})
			}

			d := &queryParamDecoder{
				values:       values,
				objectFields: fields,
				paramName:    test.Param,
				style:        test.Style,
				explode:      test.Explode,
			}

			var result []Field
			err = d.DecodeFields(func(name string, d Decoder) error {
				v, err := d.DecodeValue()
				if err != nil {
					return err
				}
				result = append(result, Field{name, v})
				return nil
			})
			require.NoError(t, err, fmt.Sprintf("Test %d", i+1))
			require.Equal(t, test.Expect, result, fmt.Sprintf("Test %d", i+1))
		}
	})
}
