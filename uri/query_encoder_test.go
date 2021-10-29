package uri

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueryEncoder(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		tests := []struct {
			Input   string
			Expect  string
			Style   QueryStyle
			Explode bool
		}{
			{
				Input:   "a",
				Expect:  "a",
				Style:   QueryStyleForm,
				Explode: true,
			},
			{
				Input:   "a",
				Expect:  "a",
				Style:   QueryStyleForm,
				Explode: false,
			},
		}

		for i, test := range tests {
			result := NewQueryEncoder(QueryEncoderConfig{
				Style:   test.Style,
				Explode: test.Explode,
			}).EncodeString(test.Input)
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
				Input:   []string{"a", "b", "c"},
				Expect:  []string{"a,b,c"},
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
			// 	Input:   []string{"a", "b", "c"},
			// 	Expect:  []string{"a%20b%20c"},
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
				Input:   []string{"a", "b", "c"},
				Expect:  []string{"a|b|c"},
				Style:   QueryStylePipeDelimited,
				Explode: false,
			},
		}

		for i, test := range tests {
			result := NewQueryEncoder(QueryEncoderConfig{
				Style:   test.Style,
				Explode: test.Explode,
			}).EncodeStrings(test.Input)
			require.Equal(t, test.Expect, result, fmt.Sprintf("Test %d", i+1))
		}
	})
}
