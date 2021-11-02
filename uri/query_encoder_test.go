package uri

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueryEncoder(t *testing.T) {
	t.Run("Value", func(t *testing.T) {
		tests := []struct {
			Param   string
			Input   string
			Expect  []string
			Style   QueryStyle
			Explode bool
		}{
			{
				Param:   "id",
				Input:   "a",
				Expect:  []string{"a"},
				Style:   QueryStyleForm,
				Explode: true,
			},
			{
				Param:   "id",
				Input:   "a",
				Expect:  []string{"a"},
				Style:   QueryStyleForm,
				Explode: false,
			},
		}

		for i, test := range tests {
			e := NewQueryEncoder(QueryEncoderConfig{
				Param:   test.Param,
				Style:   test.Style,
				Explode: test.Explode,
			})
			require.NoError(t, e.EncodeValue(test.Input))
			require.Equal(t, test.Expect, e.Result(), fmt.Sprintf("Test %d", i+1))
		}
	})

	t.Run("Array", func(t *testing.T) {
		tests := []struct {
			Param   string
			Input   []string
			Expect  []string
			Style   QueryStyle
			Explode bool
		}{
			{
				Param:   "id",
				Input:   []string{"a", "b", "c"},
				Expect:  []string{"a", "b", "c"},
				Style:   QueryStyleForm,
				Explode: true,
			},
			{
				Param:   "id",
				Input:   []string{"a", "b", "c"},
				Expect:  []string{"a,b,c"},
				Style:   QueryStyleForm,
				Explode: false,
			},
			{
				Param:   "id",
				Input:   []string{"a", "b", "c"},
				Expect:  []string{"a", "b", "c"},
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
				Input:   []string{"a", "b", "c"},
				Expect:  []string{"a", "b", "c"},
				Style:   QueryStylePipeDelimited,
				Explode: true,
			},
			{
				Param:   "id",
				Input:   []string{"a", "b", "c"},
				Expect:  []string{"a|b|c"},
				Style:   QueryStylePipeDelimited,
				Explode: false,
			},
		}

		for i, test := range tests {
			e := NewQueryEncoder(QueryEncoderConfig{
				Param:   test.Param,
				Style:   test.Style,
				Explode: test.Explode,
			})
			err := e.EncodeArray(func(e Encoder) error {
				for _, item := range test.Input {
					if err := e.EncodeValue(item); err != nil {
						return err
					}
				}
				return nil
			})
			require.NoError(t, err)
			require.Equal(t, test.Expect, e.Result(), fmt.Sprintf("Test %d", i+1))
		}
	})

	t.Run("Object", func(t *testing.T) {
		tests := []struct {
			Param   string
			Input   []Field
			Expect  []string
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
				Expect:  []string{"role=admin", "firstName=Alex"},
			},
			{
				Param: "id",
				Input: []Field{
					{"role", "admin"},
					{"firstName", "Alex"},
				},
				Style:   QueryStyleForm,
				Explode: false,
				Expect:  []string{"role,admin,firstName,Alex"},
			},
			{
				Param: "id",
				Input: []Field{
					{"role", "admin"},
					{"firstName", "Alex"},
				},
				Style:   QueryStyleDeepObject,
				Explode: true,
				Expect:  []string{"id[role]=admin", "id[firstName]=Alex"},
			},
		}

		for i, test := range tests {
			e := NewQueryEncoder(QueryEncoderConfig{
				Param:   test.Param,
				Style:   test.Style,
				Explode: test.Explode,
			})
			for _, field := range test.Input {
				err := e.EncodeField(field.Name, func(e Encoder) error {
					return e.EncodeValue(field.Value)
				})
				require.NoError(t, err)
			}
			require.Equal(t, test.Expect, e.Result(), fmt.Sprintf("Test %d", i+1))
		}
	})

}
