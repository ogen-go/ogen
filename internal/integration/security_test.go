package integration_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_security"
	"github.com/ogen-go/ogen/ogenerrors"
)

type testSecurity struct {
	basicAuth   api.BasicAuth
	bearerToken api.BearerToken
	headerKey   api.HeaderKey
	queryKey    api.QueryKey
}

func (t *testSecurity) OptionalSecurity(ctx context.Context) error {
	return nil
}

func (t *testSecurity) DisjointSecurity(ctx context.Context) error {
	return nil
}

func (t *testSecurity) IntersectSecurity(ctx context.Context) error {
	return nil
}

type tokenKey string

func (t *testSecurity) HandleBasicAuth(ctx context.Context, operationName string, v api.BasicAuth) (context.Context, error) {
	if v != t.basicAuth {
		return nil, errors.Errorf("invalid basic auth: %q and %q", v.Username, v.Password)
	}
	return context.WithValue(ctx, tokenKey("BasicAuth"), v), nil
}

func (t *testSecurity) HandleBearerToken(ctx context.Context, operationName string, v api.BearerToken) (context.Context, error) {
	if v != t.bearerToken {
		return nil, errors.Errorf("invalid token: %q", v.Token)
	}
	return context.WithValue(ctx, tokenKey("BearerToken"), v), nil
}

func (t *testSecurity) HandleHeaderKey(ctx context.Context, operationName string, v api.HeaderKey) (context.Context, error) {
	if v != t.headerKey {
		return nil, errors.Errorf("invalid api key: %q", v.APIKey)
	}
	return context.WithValue(ctx, tokenKey("HeaderKey"), v), nil
}

func (t *testSecurity) HandleQueryKey(ctx context.Context, operationName string, v api.QueryKey) (context.Context, error) {
	if v != t.queryKey {
		return nil, errors.Errorf("invalid api key: %q", v.APIKey)
	}
	return context.WithValue(ctx, tokenKey("QueryKey"), v), nil
}

type testSecuritySource struct {
	basicAuth   *api.BasicAuth
	bearerToken *api.BearerToken
	headerKey   *api.HeaderKey
	queryKey    *api.QueryKey
}

func (t *testSecuritySource) BasicAuth(ctx context.Context, operationName string) (r api.BasicAuth, _ error) {
	if v := t.basicAuth; v != nil {
		return *v, nil
	}
	return r, ogenerrors.ErrSkipClientSecurity
}

func (t *testSecuritySource) BearerToken(ctx context.Context, operationName string) (r api.BearerToken, _ error) {
	if v := t.bearerToken; v != nil {
		return *v, nil
	}
	return r, ogenerrors.ErrSkipClientSecurity
}

func (t *testSecuritySource) HeaderKey(ctx context.Context, operationName string) (r api.HeaderKey, _ error) {
	if v := t.headerKey; v != nil {
		return *v, nil
	}
	return r, ogenerrors.ErrSkipClientSecurity
}

func (t *testSecuritySource) QueryKey(ctx context.Context, operationName string) (r api.QueryKey, _ error) {
	if v := t.queryKey; v != nil {
		return *v, nil
	}
	return r, ogenerrors.ErrSkipClientSecurity
}

func TestSecurity(t *testing.T) {
	h := &testSecurity{
		basicAuth:   api.BasicAuth{Username: "username", Password: "password"},
		bearerToken: api.BearerToken{Token: "BearerToken"},
		headerKey:   api.HeaderKey{APIKey: "HeaderKey"},
		queryKey:    api.QueryKey{APIKey: "QueryKey"},
	}
	srv, err := api.NewServer(h, h)
	require.NoError(t, err)

	s := httptest.NewServer(srv)
	t.Cleanup(func() {
		s.Close()
	})

	client, err := api.NewClient(s.URL, &testSecuritySource{
		basicAuth:   &h.basicAuth,
		bearerToken: &h.bearerToken,
		headerKey:   &h.headerKey,
		queryKey:    &h.queryKey,
	}, api.WithClient(s.Client()))
	require.NoError(t, err)

	sendReq := func(t *testing.T, apiPath string, modify func(r *http.Request)) *http.Response {
		req, err := http.NewRequest(http.MethodGet, s.URL+apiPath, nil)
		require.NoError(t, err)
		if modify != nil {
			modify(req)
		}
		resp, err := s.Client().Do(req)
		require.NoError(t, err)
		// We don't care about the response body, so we can close it right away.
		require.NoError(t, resp.Body.Close())
		return resp
	}
	setQuery := func(k, v string) func(r *http.Request) {
		return func(r *http.Request) {
			q := r.URL.Query()
			q.Set(k, v)
			r.URL.RawQuery = q.Encode()
		}
	}

	t.Run("OptionalSecurity", func(t *testing.T) {
		// Empty request: okay, security is optional.
		resp := sendReq(t, "/optionalSecurity", nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		// Set "api_key" query key to invalid value.
		resp = sendReq(t, "/optionalSecurity", setQuery("api_key", "a"))
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		// Set "api_key" query key to valid value.
		resp = sendReq(t, "/optionalSecurity", setQuery("api_key", h.queryKey.APIKey))
		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.NoError(t, client.OptionalSecurity(context.Background()))
	})
	t.Run("DisjointSecurity", func(t *testing.T) {
		resp := sendReq(t, "/disjointSecurity", nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp = sendReq(t, "/disjointSecurity", setQuery("api_key", h.queryKey.APIKey))
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp = sendReq(t, "/disjointSecurity", func(r *http.Request) {
			r.SetBasicAuth(h.basicAuth.Username, h.basicAuth.Password)
			setQuery("api_key", h.queryKey.APIKey)(r)
		})
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp = sendReq(t, "/disjointSecurity", func(r *http.Request) {
			r.Header.Set("X-API-Key", h.headerKey.APIKey)
		})
		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.NoError(t, client.DisjointSecurity(context.Background()))
	})
	t.Run("IntersectSecurity", func(t *testing.T) {
		resp := sendReq(t, "/intersectSecurity", nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp = sendReq(t, "/intersectSecurity", func(r *http.Request) {
			r.Header.Set("X-API-Key", h.headerKey.APIKey)
		})
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp = sendReq(t, "/intersectSecurity", func(r *http.Request) {
			r.SetBasicAuth(h.basicAuth.Username, h.basicAuth.Password)
			r.Header.Set("X-API-Key", h.headerKey.APIKey)
		})
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp = sendReq(t, "/intersectSecurity", func(r *http.Request) {
			r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.bearerToken.Token))
			r.Header.Set("X-API-Key", h.headerKey.APIKey)
		})
		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.NoError(t, client.IntersectSecurity(context.Background()))
	})
}

func TestSecurityClientCheck(t *testing.T) {
	h := &testSecurity{
		basicAuth:   api.BasicAuth{Username: "username", Password: "password"},
		bearerToken: api.BearerToken{Token: "BearerToken"},
		headerKey:   api.HeaderKey{APIKey: "HeaderKey"},
		queryKey:    api.QueryKey{APIKey: "QueryKey"},
	}
	srv, err := api.NewServer(h, h)
	require.NoError(t, err)

	s := httptest.NewServer(srv)
	t.Cleanup(func() {
		s.Close()
	})

	type testCase struct {
		source  testSecuritySource
		wantErr bool
	}
	test := func(f func(*api.Client, context.Context) error, tts []testCase) func(t *testing.T) {
		return func(t *testing.T) {
			for i, tt := range tts {
				tt := tt
				t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
					client, err := api.NewClient(s.URL, &tt.source, api.WithClient(s.Client()))
					require.NoError(t, err)

					err = f(client, context.Background())
					if tt.wantErr {
						require.Error(t, err)
						return
					}
					require.NoError(t, err)
				})
			}
		}
	}

	t.Run("OptionalSecurity", test((*api.Client).OptionalSecurity, []testCase{
		{wantErr: false},
		{source: testSecuritySource{queryKey: &h.queryKey}, wantErr: false},
		{source: testSecuritySource{headerKey: &h.headerKey}, wantErr: false},
	}))
	t.Run("DisjointSecurity", test((*api.Client).DisjointSecurity, []testCase{
		{wantErr: true},
		{source: testSecuritySource{queryKey: &h.queryKey}, wantErr: true},

		{source: testSecuritySource{queryKey: &h.queryKey, basicAuth: &h.basicAuth}, wantErr: false},
		{source: testSecuritySource{headerKey: &h.headerKey}, wantErr: false},
	}))
	t.Run("IntersectSecurity", test((*api.Client).IntersectSecurity, []testCase{
		{wantErr: true},
		{source: testSecuritySource{queryKey: &h.queryKey}, wantErr: true},
		{source: testSecuritySource{headerKey: &h.headerKey}, wantErr: true},

		{source: testSecuritySource{headerKey: &h.headerKey, basicAuth: &h.basicAuth}, wantErr: false},
		{source: testSecuritySource{headerKey: &h.headerKey, bearerToken: &h.bearerToken}, wantErr: false},
	}))
}
