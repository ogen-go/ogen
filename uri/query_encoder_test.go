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
			Expect  string
			Style   QueryStyle
			Explode bool
		}{
			{
				Param:   "id",
				Input:   "a",
				Expect:  "a",
				Style:   QueryStyleForm,
				Explode: true,
			},
			{
				Param:   "id",
				Input:   "a",
				Expect:  "a",
				Style:   QueryStyleForm,
				Explode: false,
			},
		}

		for i, test := range tests {
			result := NewQueryEncoder(QueryEncoderConfig{
				Param:   test.Param,
				Style:   test.Style,
				Explode: test.Explode,
			}).EncodeValue(test.Input)
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
			result := NewQueryEncoder(QueryEncoderConfig{
				Param:   test.Param,
				Style:   test.Style,
				Explode: test.Explode,
			}).EncodeArray(test.Input)
			require.Equal(t, test.Expect, result, fmt.Sprintf("Test %d", i+1))
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
			result := NewQueryEncoder(QueryEncoderConfig{
				Param:   test.Param,
				Style:   test.Style,
				Explode: test.Explode,
			}).EncodeObject(test.Input)
			require.Equal(t, test.Expect, result, fmt.Sprintf("Test %d", i+1))
		}
	})

}
