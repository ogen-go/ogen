package uri

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeEscapedPath(t *testing.T) {
	tests := []struct {
		s    string
		want string
		ok   bool
	}{
		// Fast path.
		{"", "", true},
		{"/foo", "/foo", true},
		{"/foo/bar", "/foo/bar", true},
		{"/foo%00bar", "/foo%00bar", true},
		{"/foo%0Abar", "/foo%0Abar", true},
		{"/foo%20bar", "/foo%20bar", true},
		{"/foo%3Fbar", "/foo%3Fbar", true},
		{"/foo%25bar", "/foo%25bar", true},

		// Slow path.
		// Unnecessary escapes.
		{"/user/ern%61do", "/user/ernado", true},
		{"/user/ern%41do", "/user/ernAdo", true},
		// Lowercase hex digits.
		{"/foo%3fbar", "/foo%3Fbar", true},
		{"/foo%3fbar", "/foo%3Fbar", true},

		// Invalid.
		{"/foo%", "", false},
		{"/foo%3", "", false},
		{"/foo%zz", "", false},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			got, ok := NormalizeEscapedPath(tt.s)
			if !tt.ok {
				a.False(ok)
				return
			}
			a.True(ok)
			a.Equal(tt.want, got)
		})
	}
}
