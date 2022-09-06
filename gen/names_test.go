package gen

import (
	"fmt"
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

func Test_cleanRef(t *testing.T) {
	tests := []struct {
		ref  string
		want string
	}{
		{"#/components/schemas/user", "user"},
		{"#/schemas/user", "user"},
		{"#/user", "user"},
		{"user", "user"},
		{"https://example.com/foo/bar.json#/components/schemas/user", "user"},
		{"foo/bar.json#/components/schemas/user", "user"},
		{"foo/user.json", "user"},
		{"../foo/user.json", "user"},
		{"user.json", "user"},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i), func(t *testing.T) {
			require.Equal(t, tt.want, cleanRef(tt.ref))
		})
	}
}
