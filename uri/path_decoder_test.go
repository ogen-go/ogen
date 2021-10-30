package uri

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPathDecoder(t *testing.T) {
	t.Run("Value", func(t *testing.T) {
		tests := []struct {
			Param   string
			Input   string
			Expect  string
			Style   PathStyle
			Explode bool
		}{
			{
				Param:   "id",
				Input:   "foo",
				Expect:  "foo",
				Style:   PathStyleSimple,
				Explode: false,
			},
			{
				Param:   "id",
				Input:   "foo",
				Expect:  "foo",
				Style:   PathStyleSimple,
				Explode: true,
			},
			{
				Param:   "id",
				Input:   ".foo",
				Expect:  "foo",
				Style:   PathStyleLabel,
				Explode: false,
			},
			{
				Param:   "id",
				Input:   ".foo",
				Expect:  "foo",
				Style:   PathStyleLabel,
				Explode: true,
			},
			{
				Param:   "id",
				Input:   ";id=foo",
				Expect:  "foo",
				Style:   PathStyleMatrix,
				Explode: false,
			},
			{
				Param:   "id",
				Input:   ";id=foo",
				Expect:  "foo",
				Style:   PathStyleMatrix,
				Explode: true,
			},
		}
		for i, test := range tests {
			s, err := NewPathDecoder(PathDecoderConfig{
				Param:   test.Param,
				Value:   test.Input,
				Style:   test.Style,
				Explode: test.Explode,
			}).DecodeValue()

			require.NoError(t, err, fmt.Sprintf("Test %d", i+1))
			require.Equal(t, test.Expect, s, fmt.Sprintf("Test %d", i+1))
		}
	})

	t.Run("Array", func(t *testing.T) {
		tests := []struct {
			Param   string
			Input   string
			Expect  []string
			Style   PathStyle
			Explode bool
		}{
			{
				Param:   "id",
				Input:   "a,b,c",
				Expect:  []string{"a", "b", "c"},
				Style:   PathStyleSimple,
				Explode: false,
			},
			{
				Param:   "id",
				Input:   "a,b,c",
				Expect:  []string{"a", "b", "c"},
				Style:   PathStyleSimple,
				Explode: true,
			},
			{
				Param:   "id",
				Input:   ".a,b,c",
				Expect:  []string{"a", "b", "c"},
				Style:   PathStyleLabel,
				Explode: false,
			},
			{
				Param:   "id",
				Input:   ".a.b.c",
				Expect:  []string{"a", "b", "c"},
				Style:   PathStyleLabel,
				Explode: true,
			},
			{
				Param:   "id",
				Input:   ";id=a,b,c",
				Expect:  []string{"a", "b", "c"},
				Style:   PathStyleMatrix,
				Explode: false,
			},
			{
				Param:   "id",
				Input:   ";id=a;id=b;id=c",
				Expect:  []string{"a", "b", "c"},
				Style:   PathStyleMatrix,
				Explode: true,
			},
		}

		for i, test := range tests {
			s, err := NewPathDecoder(PathDecoderConfig{
				Param:   test.Param,
				Value:   test.Input,
				Style:   test.Style,
				Explode: test.Explode,
			}).DecodeArray()

			require.NoError(t, err, fmt.Sprintf("Test %d", i+1))
			require.Equal(t, test.Expect, s, fmt.Sprintf("Test %d", i+1))
		}
	})

	t.Run("Object", func(t *testing.T) {
		type field struct {
			Field string
			Value string
		}
		tests := []struct {
			Param   string
			Input   string
			Expect  []field
			Style   PathStyle
			Explode bool
		}{
			{
				Param:   "id",
				Input:   "role,admin,firstName,Alex",
				Style:   PathStyleSimple,
				Explode: false,
				Expect: []field{
					{"role", "admin"},
					{"firstName", "Alex"},
				},
			},
			{
				Param:   "id",
				Input:   "role=admin,firstName=Alex",
				Style:   PathStyleSimple,
				Explode: true,
				Expect: []field{
					{"role", "admin"},
					{"firstName", "Alex"},
				},
			},
			// {
			// 	Param:   "id",
			// 	Input:   ".role,admin,firstName,Alex",
			// 	Style:   PathStyleLabel,
			// 	Explode: false,
			// 	Expect: []field{
			// 		{"role", "admin"},
			// 		{"firstName", "Alex"},
			// 	},
			// },
			// {
			// 	Param:   "id",
			// 	Input:   ".role=admin.firstName=Alex",
			// 	Style:   PathStyleLabel,
			// 	Explode: true,
			// 	Expect: []field{
			// 		{"role", "admin"},
			// 		{"firstName", "Alex"},
			// 	},
			// },
			// {
			// 	Param:   "id",
			// 	Input:   ";id=role,admin,firstName,Alex",
			// 	Style:   PathStyleMatrix,
			// 	Explode: false,
			// 	Expect: []field{
			// 		{"role", "admin"},
			// 		{"firstName", "Alex"},
			// 	},
			// },
			// {
			// 	Param:   "id",
			// 	Input:   ";role=admin;firstName=Alex",
			// 	Style:   PathStyleMatrix,
			// 	Explode: true,
			// 	Expect: []field{
			// 		{"role", "admin"},
			// 		{"firstName", "Alex"},
			// 	},
			// },
		}

		for i, test := range tests {
			var fields []field
			err := NewPathDecoder(PathDecoderConfig{
				Param:   test.Param,
				Value:   test.Input,
				Style:   test.Style,
				Explode: test.Explode,
			}).DecodeObject(func(fieldName, value string) error {
				fields = append(fields, field{fieldName, value})
				return nil
			})
			require.NoError(t, err, fmt.Sprintf("Test %d", i+1))
			require.Equal(t, test.Expect, fields, fmt.Sprintf("Test %d", i+1))
		}
	})
}
