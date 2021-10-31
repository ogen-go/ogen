package uri

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueryDecoder(t *testing.T) {
	t.Run("Value", func(t *testing.T) {
		tests := []struct {
			Param   string
			Input   []string
			Expect  string
			Style   QueryStyle
			Explode bool
		}{
			{
				Param:   "id",
				Input:   []string{"3"},
				Expect:  "3",
				Style:   QueryStyleForm,
				Explode: true,
			},
			{
				Param:   "id",
				Input:   []string{"3"},
				Expect:  "3",
				Style:   QueryStyleForm,
				Explode: false,
			},
		}

		for i, test := range tests {
			result, err := NewQueryDecoder(QueryDecoderConfig{
				Param:   test.Param,
				Values:  test.Input,
				Style:   test.Style,
				Explode: test.Explode,
			}).Value()
			require.NoError(t, err, fmt.Sprintf("Test %d", i+1))
			require.Equal(t, test.Expect, result, fmt.Sprintf("Test %d", i+1))
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
				Input:   []string{"a,b,c"},
				Expect:  []string{"a", "b", "c"},
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
			// 	Input:   []string{"a%20b%20c"},
			// 	Expect:  []string{"a", "b", "c"},
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
				Input:   []string{"a|b|c"},
				Expect:  []string{"a", "b", "c"},
				Style:   QueryStylePipeDelimited,
				Explode: false,
			},
		}

		for i, test := range tests {
			d := NewQueryDecoder(QueryDecoderConfig{
				Param:   test.Param,
				Values:  test.Input,
				Style:   test.Style,
				Explode: test.Explode,
			})

			var items []string
			err := d.Array(func(d Decoder) error {
				item, err := d.Value()
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
			Input   []string
			Expect  []Field
			Style   QueryStyle
			Explode bool
		}{
			{
				Param:   "id",
				Input:   []string{"role=admin", "firstName=Alex"},
				Style:   QueryStyleForm,
				Explode: true,
				Expect: []Field{
					{"role", "admin"},
					{"firstName", "Alex"},
				},
			},
			{
				Param:   "id",
				Input:   []string{"id=role,admin,firstName,Alex"},
				Style:   QueryStyleForm,
				Explode: false,
				Expect: []Field{
					{"role", "admin"},
					{"firstName", "Alex"},
				},
			},
			{
				Param:   "id",
				Input:   []string{"id[role]=admin", "id[firstName]=Alex"},
				Style:   QueryStyleDeepObject,
				Explode: true,
				Expect: []Field{
					{"role", "admin"},
					{"firstName", "Alex"},
				},
			},
		}

		for i, test := range tests {
			var fields []Field
			d := NewQueryDecoder(QueryDecoderConfig{
				Param:   test.Param,
				Values:  test.Input,
				Style:   test.Style,
				Explode: test.Explode,
			})

			err := d.Fields(func(name string, d Decoder) error {
				v, err := d.Value()
				if err != nil {
					return err
				}
				fields = append(fields, Field{name, v})
				return nil
			})
			require.NoError(t, err, fmt.Sprintf("Test %d", i+1))
			require.Equal(t, test.Expect, fields, fmt.Sprintf("Test %d", i+1))
		}
	})
}
