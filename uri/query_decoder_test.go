package uri

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueryDecoder(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		tests := []struct {
			Input   []string
			Expect  string
			Style   QueryStyle
			Explode bool
		}{
			{
				Input:   []string{"3"},
				Expect:  "3",
				Style:   QueryStyleForm,
				Explode: true,
			},
			{
				Input:   []string{"3"},
				Expect:  "3",
				Style:   QueryStyleForm,
				Explode: false,
			},
		}

		for i, test := range tests {
			result, err := NewQueryDecoder(QueryDecoderConfig{
				Values:  test.Input,
				Style:   test.Style,
				Explode: test.Explode,
			}).DecodeString()
			require.NoError(t, err, fmt.Sprintf("Test %d", i+1))
			require.Equal(t, test.Expect, result, fmt.Sprintf("Test %d", i+1))
		}
	})

	t.Run("StringArray", func(t *testing.T) {
		tests := []struct {
			Input   []string
			Expect  []string
			Style   QueryStyle
			Explode bool
		}{
			{
				Input:   []string{"a", "b", "c"},
				Expect:  []string{"a", "b", "c"},
				Style:   QueryStyleForm,
				Explode: true,
			},
			{
				Input:   []string{"a,b,c"},
				Expect:  []string{"a", "b", "c"},
				Style:   QueryStyleForm,
				Explode: false,
			},
			{
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
				Input:   []string{"a", "b", "c"},
				Expect:  []string{"a", "b", "c"},
				Style:   QueryStylePipeDelimited,
				Explode: true,
			},
			{
				Input:   []string{"a|b|c"},
				Expect:  []string{"a", "b", "c"},
				Style:   QueryStylePipeDelimited,
				Explode: false,
			},
		}

		for i, test := range tests {
			result, err := NewQueryDecoder(QueryDecoderConfig{
				Values:  test.Input,
				Style:   test.Style,
				Explode: test.Explode,
			}).DecodeStrings()
			require.NoError(t, err, fmt.Sprintf("Test %d", i+1))
			require.Equal(t, test.Expect, result, fmt.Sprintf("Test %d", i+1))
		}
	})
}
