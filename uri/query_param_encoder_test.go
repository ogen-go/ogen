package uri

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueryParamEncoder(t *testing.T) {
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
				Input:   "a",
				Expect:  "id=a",
				Style:   QueryStyleForm,
				Explode: true,
			},
			{
				Param:   "id",
				Input:   "a",
				Expect:  "id=a",
				Style:   QueryStyleForm,
				Explode: false,
			},
		}

		for i, test := range tests {
			e := queryParamEncoder{
				receiver:  newReceiver(),
				paramName: test.Param,
				style:     test.Style,
				explode:   test.Explode,
				values:    make(url.Values),
			}
			require.NoError(t, e.EncodeValue(test.Input))
			require.NoError(t, e.serialize())
			require.Equal(t, test.Expect, e.values.Encode(), fmt.Sprintf("Test %d", i+1))
		}
	})

	t.Run("Array", func(t *testing.T) {
		tests := []struct {
			Param   string
			Input   []string
			Expect  string
			Style   QueryStyle
			Explode bool
		}{
			{
				Param:   "id",
				Input:   []string{"3", "4", "5"},
				Expect:  "id=3&id=4&id=5",
				Style:   QueryStyleForm,
				Explode: true,
			},
			{
				Param:   "id",
				Input:   []string{"3", "4", "5"},
				Expect:  "id=3%2C4%2C5",
				Style:   QueryStyleForm,
				Explode: false,
			},
			{
				Param:   "id",
				Input:   []string{"3", "4", "5"},
				Expect:  "id=3&id=4&id=5",
				Style:   QueryStyleSpaceDelimited,
				Explode: true,
			},
			// {
			// 	Input:   []string{"a", "b", "c"},
			// 	Expect:  []string{"a%20b%20c"},
			// 	Style:   QueryStyleSpaceDelimited,
			// 	Explode: false,
			// },
			{
				Param:   "id",
				Input:   []string{"3", "4", "5"},
				Expect:  "id=3&id=4&id=5",
				Style:   QueryStylePipeDelimited,
				Explode: true,
			},
			{
				Param:   "id",
				Input:   []string{"3", "4", "5"},
				Expect:  "id=3%7C4%7C5",
				Style:   QueryStylePipeDelimited,
				Explode: false,
			},
		}

		for i, test := range tests {
			e := queryParamEncoder{
				receiver:  newReceiver(),
				paramName: test.Param,
				style:     test.Style,
				explode:   test.Explode,
				values:    make(url.Values),
			}
			err := e.EncodeArray(func(e Encoder) error {
				for _, item := range test.Input {
					if err := e.EncodeValue(item); err != nil {
						return err
					}
				}
				return nil
			})
			require.NoError(t, err)
			require.NoError(t, e.serialize())
			require.Equal(t, test.Expect, e.values.Encode(), fmt.Sprintf("Test %d", i+1))
		}
	})

	t.Run("Object", func(t *testing.T) {
		tests := []struct {
			Param   string
			Input   []Field
			Expect  string
			Style   QueryStyle
			Explode bool
		}{
			{
				Param: "id",
				Input: []Field{
					{"role", "admin"},
					{"firstName", "Alex"},
				},
				Style:   QueryStyleForm,
				Explode: true,
				Expect:  "firstName=Alex&role=admin",
			},
			{
				Param: "id",
				Input: []Field{
					{"role", "admin"},
					{"firstName", "Alex"},
				},
				Style:   QueryStyleForm,
				Explode: false,
				Expect:  "id=role%2Cadmin%2CfirstName%2CAlex",
			},
			{
				Param: "id",
				Input: []Field{
					{"role", "admin"},
					{"firstName", "Alex"},
				},
				Style:   QueryStyleDeepObject,
				Explode: true,
				Expect:  "id%5BfirstName%5D=Alex&id%5Brole%5D=admin",
			},
		}

		for i, test := range tests {
			e := queryParamEncoder{
				receiver:  newReceiver(),
				paramName: test.Param,
				style:     test.Style,
				explode:   test.Explode,
				values:    make(url.Values),
			}
			for _, field := range test.Input {
				err := e.EncodeField(field.Name, func(e Encoder) error {
					return e.EncodeValue(field.Value)
				})
				require.NoError(t, err)
			}
			require.NoError(t, e.serialize())
			require.Equal(t, test.Expect, e.values.Encode(), fmt.Sprintf("Test %d", i+1))
		}
	})
}
