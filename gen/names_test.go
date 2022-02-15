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
		Error   bool
	}{
		{"user_id", "UserID", false, false},
		{"userId", "UserId", false, false},
		{"foo+bar", "FooPlusBar", true, false},
		{"foo+bar", "FooBar", false, false},
		{"+1", "Plus1", true, false},
	}

	for _, test := range tests {
		out, err := (&nameGen{
			src:          []rune(test.Input),
			allowSpecial: test.AllowMP,
		}).generate()
		require.NoError(t, err)
		require.Equal(t, test.Expect, out)
	}
}
