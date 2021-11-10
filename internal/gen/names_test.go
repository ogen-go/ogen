package gen

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNames(t *testing.T) {
	tests := []struct {
		Input   string
		Expect  string
		AllowMP bool
	}{
		{"user_id", "UserID", false},
		{"userId", "UserId", false},
		{"foo+bar", "FooPlusBar", true},
		{"foo+bar", "FooBar", false},
		{"+1", "Plus1", true},
	}

	for _, test := range tests {
		out := (&nameGen{
			src:          []rune(test.Input),
			allowSpecial: test.AllowMP,
		}).generate()
		require.Equal(t, test.Expect, out)
	}
}
