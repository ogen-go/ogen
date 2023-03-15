package uri

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkCloneURL(b *testing.B) {
	u, err := url.Parse("https://go.dev")
	require.NoError(b, err)
	var uCloned *url.URL

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		uCloned = Clone(u)
	}

	if uCloned.Host != "go.dev" {
		b.Fatal(uCloned)
	}
}

func TestAddPathParts(t *testing.T) {
	const (
		exampleAPI  = "https://example.com/api/v1"
		emptyPath   = "https://example.com"
		escapedBase = "https://example.com/api/v%2F1"
	)
	tests := []struct {
		Root    string
		Parts   []string
		Path    string
		RawPath string
		URL     string
	}{
		// Plain base path.
		{
			Root:    exampleAPI,
			Parts:   []string{"/foo/baz/bar"},
			Path:    "/api/v1/foo/baz/bar",
			RawPath: "",
			URL:     "https://example.com/api/v1/foo/baz/bar",
		},
		{
			Root:    exampleAPI,
			Parts:   []string{"/repo/", "ogen-go", "/", "ogen", "/issues"},
			Path:    "/api/v1/repo/ogen-go/ogen/issues",
			RawPath: "",
			URL:     "https://example.com/api/v1/repo/ogen-go/ogen/issues",
		},
		{
			Root:    exampleAPI,
			Parts:   []string{"/repo/", "ogen%25go", "/", "ogen"},
			Path:    "/api/v1/repo/ogen%go/ogen",
			RawPath: "/api/v1/repo/ogen%25go/ogen",
			URL:     "https://example.com/api/v1/repo/ogen%25go/ogen",
		},

		// Empty base path.
		{
			Root:    emptyPath,
			Parts:   []string{"/repo/", "ogen%25go", "/", "ogen"},
			Path:    "/repo/ogen%go/ogen",
			RawPath: "/repo/ogen%25go/ogen",
			URL:     "https://example.com/repo/ogen%25go/ogen",
		},

		// Escaped base path.
		{
			Root:    escapedBase,
			Parts:   []string{"/foo/baz/bar"},
			Path:    "/api/v/1/foo/baz/bar",
			RawPath: "/api/v%2F1/foo/baz/bar",
			URL:     "https://example.com/api/v%2F1/foo/baz/bar",
		},
		{
			Root:    escapedBase,
			Parts:   []string{"/repo/", "ogen-go", "/", "ogen", "/issues"},
			Path:    "/api/v/1/repo/ogen-go/ogen/issues",
			RawPath: "/api/v%2F1/repo/ogen-go/ogen/issues",
			URL:     "https://example.com/api/v%2F1/repo/ogen-go/ogen/issues",
		},
		{
			Root:    escapedBase,
			Parts:   []string{"/repo/", "ogen%25go", "/", "ogen"},
			Path:    "/api/v/1/repo/ogen%go/ogen",
			RawPath: "/api/v%2F1/repo/ogen%25go/ogen",
			URL:     "https://example.com/api/v%2F1/repo/ogen%25go/ogen",
		},
	}

	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			u, err := url.Parse(tt.Root)
			a.NoError(err)

			AddPathParts(u, tt.Parts...)
			a.Equal(tt.Path, u.Path)
			a.Equal(tt.RawPath, u.RawPath)
			a.Equal(tt.URL, u.String())
		})
	}
}
