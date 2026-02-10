package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_cors"
)

type corsTestServer struct {
	api.UnimplementedHandler
}

func (corsTestServer) FooPost(ctx context.Context, req api.FooPostReq) (r *api.FooPostOK, _ error) {
	return &api.FooPostOK{Location: "/foo"}, nil
}

var _ api.SecurityHandler = corsTestServer{}

func (corsTestServer) HandleBearerToken(
	ctx context.Context,
	operationName api.OperationName,
	t api.BearerToken,
) (context.Context, error) {
	return ctx, nil
}

func (corsTestServer) HandleHeaderKey(
	ctx context.Context,
	operationName api.OperationName,
	t api.HeaderKey,
) (context.Context, error) {
	return ctx, nil
}

func TestMethodNotAllowed(t *testing.T) {
	srv, err := api.NewServer(corsTestServer{}, corsTestServer{})
	require.NoError(t, err)

	s := httptest.NewServer(srv)
	defer s.Close()

	t.Run("OPTIONS", func(t *testing.T) {
		tests := []struct {
			method  string
			headers string
		}{
			{"", ""},
			{"get", "Content-Length"},
			{"post", "Authorization,Content-Type"},
			{"patch", "Authorization,Content-Type,X-Api-Key"},
		}
		for _, tt := range tests {
			req, err := http.NewRequest("OPTIONS", s.URL+"/foo", http.NoBody)
			require.NoError(t, err)

			if tt.method != "" {
				req.Header.Set("Access-Control-Request-Method", tt.method)
			}

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())

			allowedMethods := resp.Header.Get("Access-Control-Allow-Methods")
			require.Equal(t, "GET,PATCH,POST", allowedMethods)

			allowedHeaders := resp.Header.Get("Access-Control-Allow-Headers")
			require.Equal(t, tt.headers, allowedHeaders)

			acceptPost := resp.Header.Get("Accept-Post")
			require.Equal(t, "application/json,text/plain", acceptPost)

			acceptPatch := resp.Header.Get("Accept-Patch")
			require.Equal(t, "application/sdp", acceptPatch)
		}
	})

	t.Run("POST", func(t *testing.T) {
		req, err := http.NewRequest("POST", s.URL+"/foo", strings.NewReader("foo"))
		require.NoError(t, err)

		req.Header.Set("Content-Type", "text/plain")
		req.Header.Set("Authorization", "Bearer foo")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())

		require.Equal(t, 200, resp.StatusCode)

		exposedHeaders := resp.Header.Get("Access-Control-Expose-Headers")
		require.Equal(t, "Location", exposedHeaders)

		location := resp.Header.Get("Location")
		require.Equal(t, "/foo", location)
	})

	t.Run("DELETE", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", s.URL+"/foo", http.NoBody)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())

		allow := resp.Header.Get("Allow")
		require.Equal(t, "GET,PATCH,POST", allow)
	})
}
