package uri

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func FuzzCookieEscapeUnescape(f *testing.F) {
	f.Add("foobar")
	f.Add("\xad")
	f.Add("`\"';,./<>?[]{}\\|~!@#$%^&*()_+-=")
	for _, c := range cookieEscapeChars {
		f.Add(string(c))
	}

	f.Fuzz(func(t *testing.T, s string) {
		escaped := escapeCookie(s)

		unescaped, ok := unescapeCookie(escaped)
		require.True(t, ok)
		require.Equal(t, s, unescaped)
	})
}

func Test_unescapeCookie(t *testing.T) {
	tests := []struct {
		input  string
		want   string
		wantOk bool
	}{
		{"", "", true},
		{"foobar", "foobar", true},
		{"%00", "\x00", true},
		{"%0a", "\n", true},
		{"%0A", "\n", true},

		{"%", "", false},
		{"%0", "", false},
		{"%a", "", false},
		{"%A", "", false},
		{"%0j", "", false},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			got, ok := unescapeCookie(tt.input)
			if !tt.wantOk {
				require.False(t, ok)
				return
			}

			require.True(t, ok)
			require.Equal(t, tt.want, got)
		})
	}
}
