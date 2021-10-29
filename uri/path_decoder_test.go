package uri

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPathDecoder(t *testing.T) {
	t.Run("String", func(t *testing.T) {
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
			}).DecodeString()

			require.NoError(t, err, fmt.Sprintf("Test %d", i+1))
			require.Equal(t, test.Expect, s, fmt.Sprintf("Test %d", i+1))
		}
	})

	t.Run("StringArray", func(t *testing.T) {
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
			}).DecodeStrings()

			require.NoError(t, err, fmt.Sprintf("Test %d", i+1))
			require.Equal(t, test.Expect, s, fmt.Sprintf("Test %d", i+1))
		}
	})
}
