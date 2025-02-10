package integration

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	ht "github.com/ogen-go/ogen/http"
	api "github.com/ogen-go/ogen/internal/integration/test_client_options"

	"github.com/stretchr/testify/require"
)

type clientOptionsHandler struct{}

// Foo implements api.Handler.
func (c *clientOptionsHandler) Foo(ctx context.Context, params api.FooParams) (string, error) {
	return params.Body, nil
}

var _ api.Handler = (*clientOptionsHandler)(nil)

func TestClientOptions(t *testing.T) {
	h, err := api.NewServer(&clientOptionsHandler{})
	require.NoError(t, err)

	s := httptest.NewServer(h)
	defer s.Close()

	t.Run("WithRequestClient", func(t *testing.T) {
		ctx := context.Background()

		c, err := api.NewClient(s.URL)
		require.NoError(t, err)

		op := api.WithRequestClient(new(testFaultyClient))
		_, err = c.Foo(ctx, api.FooParams{Body: "test"}, op)
		require.ErrorContains(t, err, `test faulty client`)
	})
	t.Run("WithServerURL", func(t *testing.T) {
		ctx := context.Background()

		c, err := api.NewClient(`http://completly-wrong-url.foo`)
		require.NoError(t, err)

		u, err := url.Parse(s.URL)
		require.NoError(t, err)

		op := api.WithServerURL(u)
		resp, err := c.Foo(ctx, api.FooParams{Body: "test"}, op)
		require.NoError(t, err)
		require.Equal(t, "test", resp)
	})
	t.Run("WithEditRequest", func(t *testing.T) {
		ctx := context.Background()

		c, err := api.NewClient(s.URL)
		require.NoError(t, err)

		op := api.WithEditRequest(func(req *http.Request) error {
			q := req.URL.Query()
			q.Set("body", "request-override")
			req.URL.RawQuery = q.Encode()
			return nil
		})
		resp, err := c.Foo(ctx, api.FooParams{Body: "test"}, op)
		require.NoError(t, err)
		require.Equal(t, "request-override", resp)

		op = api.WithEditRequest(func(*http.Request) error {
			return errors.New("request editor error")
		})
		_, err = c.Foo(ctx, api.FooParams{Body: "test"}, op)
		require.ErrorContains(t, err, `request editor error`)
	})
	t.Run("WithEditResponse", func(t *testing.T) {
		ctx := context.Background()

		c, err := api.NewClient(s.URL)
		require.NoError(t, err)

		op := api.WithEditResponse(func(resp *http.Response) error {
			resp.Body = io.NopCloser(strings.NewReader(`"response-override"`))
			return nil
		})
		resp, err := c.Foo(ctx, api.FooParams{Body: "test"}, op)
		require.NoError(t, err)
		require.Equal(t, "response-override", resp)

		op = api.WithEditResponse(func(*http.Response) error {
			return errors.New("response editor error")
		})
		_, err = c.Foo(ctx, api.FooParams{Body: "test"}, op)
		require.ErrorContains(t, err, `response editor error`)
	})
}

type testFaultyClient struct{}

var _ ht.Client = (*testFaultyClient)(nil)

func (f *testFaultyClient) Do(req *http.Request) (*http.Response, error) {
	return nil, errors.New("test faulty client")
}
