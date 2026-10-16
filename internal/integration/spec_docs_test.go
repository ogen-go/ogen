package integration

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
	api "github.com/ogen-go/ogen/internal/integration/test_spec_docs"
)

type specDocsServer struct {
	api.UnimplementedHandler
}

func (specDocsServer) ProbeLiveness(ctx context.Context) (string, error) {
	return "ok", nil
}

func specDocsGet(t *testing.T, c *http.Client, method, u string, hdr http.Header) (resp *http.Response, body string) {
	t.Helper()

	req, err := http.NewRequest(method, u, http.NoBody)
	require.NoError(t, err)
	for k, v := range hdr {
		req.Header[k] = v
	}

	resp, err = c.Do(req)
	require.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp, string(data)
}

func TestSpecDocs(t *testing.T) {
	newServer := func(t *testing.T, opts ...api.ServerOption) (*httptest.Server, *http.Client) {
		srv, err := api.NewServer(specDocsServer{}, opts...)
		require.NoError(t, err)
		s := httptest.NewServer(srv)
		t.Cleanup(s.Close)

		c := s.Client()
		c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
		return s, c
	}

	t.Run("Embedded", func(t *testing.T) {
		spec, err := ogen.Parse(api.SpecJSON())
		require.NoError(t, err)
		require.Equal(t, "title", spec.Info.Title)
		require.Contains(t, spec.Paths, "/healthz")

		spec, err = ogen.Parse(api.SpecYAML())
		require.NoError(t, err)
		require.Equal(t, "title", spec.Info.Title)
	})

	t.Run("WithDocs", func(t *testing.T) {
		notFound := 0
		s, c := newServer(t,
			api.WithDocs("/docs"),
			api.WithNotFound(func(w http.ResponseWriter, r *http.Request) {
				notFound++
				w.WriteHeader(http.StatusNotFound)
			}),
		)

		t.Run("JSON", func(t *testing.T) {
			resp, body := specDocsGet(t, c, http.MethodGet, s.URL+"/docs/openapi.json", nil)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			require.Equal(t, "application/json", resp.Header.Get("Content-Type"))
			require.Equal(t, string(api.SpecJSON()), body)

			etag := resp.Header.Get("ETag")
			require.NotEmpty(t, etag)
			resp, _ = specDocsGet(t, c, http.MethodGet, s.URL+"/docs/openapi.json", http.Header{
				"If-None-Match": []string{etag},
			})
			require.Equal(t, http.StatusNotModified, resp.StatusCode)
		})
		t.Run("YAML", func(t *testing.T) {
			resp, body := specDocsGet(t, c, http.MethodGet, s.URL+"/docs/openapi.yaml", nil)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			require.Equal(t, "application/yaml; charset=utf-8", resp.Header.Get("Content-Type"))
			require.Equal(t, string(api.SpecYAML()), body)
		})
		t.Run("Head", func(t *testing.T) {
			resp, body := specDocsGet(t, c, http.MethodHead, s.URL+"/docs/openapi.json", nil)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			require.Empty(t, body)
		})
		t.Run("MethodNotAllowed", func(t *testing.T) {
			resp, _ := specDocsGet(t, c, http.MethodPost, s.URL+"/docs/openapi.json", nil)
			require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
			require.Equal(t, "GET, HEAD", resp.Header.Get("Allow"))
		})
		t.Run("Redirect", func(t *testing.T) {
			resp, _ := specDocsGet(t, c, http.MethodGet, s.URL+"/docs?foo=bar", nil)
			require.Equal(t, http.StatusMovedPermanently, resp.StatusCode)
			require.Equal(t, "/docs/?foo=bar", resp.Header.Get("Location"))
		})
		t.Run("Index", func(t *testing.T) {
			resp, body := specDocsGet(t, c, http.MethodGet, s.URL+"/docs/", nil)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			require.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))
			require.Contains(t, body, "swagger-ui-bundle.js")
			require.Contains(t, body, `"openapi.json"`)
			// Default title is the API title.
			require.Contains(t, body, "<title>title</title>")
		})
		t.Run("Assets", func(t *testing.T) {
			resp, body := specDocsGet(t, c, http.MethodGet, s.URL+"/docs/swagger-ui.css", nil)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			require.Equal(t, "text/css; charset=utf-8", resp.Header.Get("Content-Type"))
			require.NotEmpty(t, body)
		})
		t.Run("NotFound", func(t *testing.T) {
			before := notFound
			for _, p := range []string{"/docs/missing", "/docsx", "/docs/openapi.json/x"} {
				resp, _ := specDocsGet(t, c, http.MethodGet, s.URL+p, nil)
				require.Equal(t, http.StatusNotFound, resp.StatusCode, p)
			}
			require.Equal(t, before+3, notFound, "custom NotFound handler must be used")
		})
		t.Run("API", func(t *testing.T) {
			resp, body := specDocsGet(t, c, http.MethodGet, s.URL+"/healthz", nil)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			require.Equal(t, `"ok"`, strings.TrimSpace(body))
		})
		t.Run("FindRoute", func(t *testing.T) {
			srv, err := api.NewServer(specDocsServer{}, api.WithDocs("/docs"))
			require.NoError(t, err)
			_, ok := srv.FindRoute(http.MethodGet, "/docs/openapi.json")
			require.False(t, ok)
			r, ok := srv.FindRoute(http.MethodGet, "/healthz")
			require.True(t, ok)
			require.Equal(t, "ProbeLiveness", r.Name())
		})
	})

	t.Run("Options", func(t *testing.T) {
		s, c := newServer(t, api.WithDocs("docs/",
			api.WithDocsTitle("Custom title"),
			api.WithDocsAssetsURL("https://cdn.example.com/swagger-ui-dist@5/"),
		))

		resp, body := specDocsGet(t, c, http.MethodGet, s.URL+"/docs/", nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Contains(t, body, "<title>Custom title</title>")
		require.Contains(t, body, `src="https://cdn.example.com/swagger-ui-dist@5/swagger-ui-bundle.js"`)

		// Assets are not served when AssetsURL is set.
		resp, _ = specDocsGet(t, c, http.MethodGet, s.URL+"/docs/swagger-ui.css", nil)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("PathPrefix", func(t *testing.T) {
		s, c := newServer(t, api.WithPathPrefix("/api"), api.WithDocs("/docs"))

		resp, _ := specDocsGet(t, c, http.MethodGet, s.URL+"/api/docs/openapi.json", nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, _ = specDocsGet(t, c, http.MethodGet, s.URL+"/api/docs", nil)
		require.Equal(t, http.StatusMovedPermanently, resp.StatusCode)
		require.Equal(t, "/api/docs/", resp.Header.Get("Location"))

		resp, _ = specDocsGet(t, c, http.MethodGet, s.URL+"/docs/openapi.json", nil)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)

		resp, _ = specDocsGet(t, c, http.MethodGet, s.URL+"/api/healthz", nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Root", func(t *testing.T) {
		s, c := newServer(t, api.WithDocs(""))

		resp, _ := specDocsGet(t, c, http.MethodGet, s.URL+"/openapi.json", nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, body := specDocsGet(t, c, http.MethodGet, s.URL+"/", nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Contains(t, body, "swagger-ui-bundle.js")

		resp, _ = specDocsGet(t, c, http.MethodGet, s.URL+"/healthz", nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Disabled", func(t *testing.T) {
		s, c := newServer(t)

		resp, _ := specDocsGet(t, c, http.MethodGet, s.URL+"/docs/openapi.json", nil)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Standalone", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.Handle("/spec/", http.StripPrefix("/spec", api.SpecHandler()))
		mux.Handle("/ui/", http.StripPrefix("/ui", api.SwaggerUIHandler()))
		s := httptest.NewServer(mux)
		defer s.Close()
		c := s.Client()

		resp, body := specDocsGet(t, c, http.MethodGet, s.URL+"/spec/openapi.json", nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, string(api.SpecJSON()), body)

		// SpecHandler does not serve Swagger UI.
		resp, _ = specDocsGet(t, c, http.MethodGet, s.URL+"/spec/", nil)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)

		resp, body = specDocsGet(t, c, http.MethodGet, s.URL+"/ui/", nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Contains(t, body, "swagger-ui-bundle.js")

		resp, _ = specDocsGet(t, c, http.MethodGet, s.URL+"/ui/openapi.yaml", nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, _ = specDocsGet(t, c, http.MethodGet, s.URL+"/ui/swagger-ui-bundle.js", nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
