package gen

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/jsonschema"
)

func TestNames(t *testing.T) {
	tests := []struct {
		Input       string
		Expect      string
		AllowMP     bool
		Initialisms bool
		Error       bool
	}{
		{"user_id", "UserID", false, false, false},
		{"userId", "UserId", false, false, false},
		{"foo+bar", "FooPlusBar", true, false, false},
		{"foo+bar", "FooBar", false, false, false},
		{"+1", "Plus1", true, false, false},

		// NamingCamelInitialisms feature: lower->upper transitions inside a
		// camelCase token are treated as word boundaries, so the initialism
		// rules fire on camelCase input the same way they do on snake_case.
		{"userId", "UserID", false, true, false},
		{"orderId", "OrderID", false, true, false},
		{"apiKey", "APIKey", false, true, false},
		{"httpUrl", "HTTPURL", false, true, false},
		{"user_id", "UserID", false, true, false}, // snake_case still works
		// Non-initialism words and acronym runs are left unchanged.
		{"fooBar", "FooBar", false, true, false},
		{"parseURL", "ParseURL", false, true, false},
		{"HTTPServer", "HTTPServer", false, true, false},
	}

	for _, test := range tests {
		out, err := (&nameGen{
			src:          []rune(test.Input),
			allowSpecial: test.AllowMP,
			initialisms:  test.Initialisms,
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
			u, err := url.Parse(tt.ref)
			require.NoError(t, err)

			var ref jsonschema.Ref
			ref.FromURL(u)

			require.Equal(t, tt.want, cleanRef(ref))
		})
	}
}
