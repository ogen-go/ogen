package swaggerui

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	h := New(Config{SpecURL: "openapi.json", Title: "My API"})
	s := httptest.NewServer(h)
	defer s.Close()

	get := func(t *testing.T, method, p string) *http.Response {
		req, err := http.NewRequest(method, s.URL+p, http.NoBody)
		require.NoError(t, err)
		resp, err := s.Client().Do(req)
		require.NoError(t, err)
		t.Cleanup(func() { _ = resp.Body.Close() })
		return resp
	}
	body := func(t *testing.T, resp *http.Response) string {
		var buf [1 << 16]byte
		n, _ := resp.Body.Read(buf[:])
		return string(buf[:n])
	}

	t.Run("Index", func(t *testing.T) {
		for _, p := range []string{"", "/"} {
			resp := get(t, http.MethodGet, p)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			require.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))
			b := body(t, resp)
			require.Contains(t, b, `<title>My API</title>`)
			require.Contains(t, b, `"openapi.json"`)
			require.Contains(t, b, `href="swagger-ui.css"`)
			require.Contains(t, b, `src="swagger-ui-bundle.js"`)
			require.Contains(t, b, `"oauth2-redirect.html"`)
		}
	})
	t.Run("Head", func(t *testing.T) {
		resp := get(t, http.MethodHead, "/")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Empty(t, body(t, resp))
	})
	t.Run("Assets", func(t *testing.T) {
		for _, tt := range []struct {
			name, contentType string
		}{
			{"swagger-ui.css", "text/css; charset=utf-8"},
			{"swagger-ui-bundle.js", "text/javascript; charset=utf-8"},
			{"swagger-ui-standalone-preset.js", "text/javascript; charset=utf-8"},
			{"favicon-32x32.png", "image/png"},
			{"oauth2-redirect.html", "text/html; charset=utf-8"},
		} {
			resp := get(t, http.MethodGet, "/"+tt.name)
			require.Equal(t, http.StatusOK, resp.StatusCode, tt.name)
			require.Equal(t, tt.contentType, resp.Header.Get("Content-Type"), tt.name)
			require.NotEmpty(t, resp.Header.Get("ETag"), tt.name)
			require.True(t, h.Match("/"+tt.name), tt.name)
		}
	})
	t.Run("NotFound", func(t *testing.T) {
		for _, p := range []string{"/nope", "/dist/swagger-ui.css", "/../swaggerui.go", "/swagger-ui.css/"} {
			resp := get(t, http.MethodGet, p)
			require.Equal(t, http.StatusNotFound, resp.StatusCode, p)
			require.False(t, h.Match(p), p)
		}
	})
	t.Run("MethodNotAllowed", func(t *testing.T) {
		resp := get(t, http.MethodPost, "/")
		require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		require.Equal(t, "GET, HEAD", resp.Header.Get("Allow"))
	})
}

func TestHandlerAssetsURL(t *testing.T) {
	h := New(Config{SpecURL: "/api/openapi.json", AssetsURL: "https://cdn.example.com/swagger-ui-dist@5"})
	s := httptest.NewServer(h)
	defer s.Close()

	resp, err := s.Client().Get(s.URL + "/")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var buf [1 << 16]byte
	n, _ := resp.Body.Read(buf[:])
	b := string(buf[:n])
	require.Contains(t, b, `<title>Swagger UI</title>`)
	require.Contains(t, b, `href="https://cdn.example.com/swagger-ui-dist@5/swagger-ui.css"`)
	require.Contains(t, b, `"https://cdn.example.com/swagger-ui-dist@5/oauth2-redirect.html"`)
	require.Contains(t, b, `"/api/openapi.json"`)

	// Assets are not served in this mode.
	require.True(t, h.Match("/"))
	require.False(t, h.Match("/swagger-ui.css"))
	resp2, err := s.Client().Get(s.URL + "/swagger-ui.css")
	require.NoError(t, err)
	defer func() { _ = resp2.Body.Close() }()
	require.Equal(t, http.StatusNotFound, resp2.StatusCode)
}

func TestAssets(t *testing.T) {
	require.NotEmpty(t, Version)
	for _, name := range []string{
		"LICENSE",
		"NOTICE",
		"swagger-ui.css",
		"swagger-ui-bundle.js",
		"swagger-ui-standalone-preset.js",
		"oauth2-redirect.html",
		"oauth2-redirect.js",
		"favicon-16x16.png",
		"favicon-32x32.png",
	} {
		st, err := fs.Stat(Assets(), name)
		require.NoError(t, err, name)
		require.NotZero(t, st.Size(), name)
	}
}
