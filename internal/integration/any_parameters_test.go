package integration_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_any_parameters"
)

type testAnyParameters struct{}

func (t testAnyParameters) AnyParams(ctx context.Context, params api.AnyParamsParams) (*api.AnyParamsOK, error) {
	// All parameters should be received as strings (or empty string if not provided)
	resp := &api.AnyParamsOK{
		Echo: api.AnyParamsOKEcho{
			PathParam: params.PathParam.(string),
		},
	}

	if params.QueryParam != nil {
		qp := params.QueryParam.(string)
		resp.Echo.QueryParam.SetTo(qp)
	}

	if params.XHeaderParam != nil {
		hp := params.XHeaderParam.(string)
		resp.Echo.HeaderParam.SetTo(hp)
	}

	if params.CookieParam != nil {
		cp := params.CookieParam.(string)
		resp.Echo.CookieParam.SetTo(cp)
	}

	return resp, nil
}

func (t testAnyParameters) AnyParamsRequired(ctx context.Context, params api.AnyParamsRequiredParams) (*api.AnyParamsRequiredOK, error) {
	// Verify all required parameters are received as strings
	_ = params.PathParam.(string)
	_ = params.QueryParam.(string)
	_ = params.XHeaderParam.(string)

	return &api.AnyParamsRequiredOK{
		Received: true,
	}, nil
}

func (t testAnyParameters) AnyArrayParam(ctx context.Context, params api.AnyArrayParamParams) (*api.AnyArrayParamOK, error) {
	// Each item in the array should be a string
	count := len(params.Items)

	// Verify all items are strings
	for _, item := range params.Items {
		_ = item.(string)
	}

	return &api.AnyArrayParamOK{
		Count: api.NewOptInt(count),
	}, nil
}

func TestAnyParameters(t *testing.T) {
	ctx := context.Background()
	h, err := api.NewServer(testAnyParameters{})
	require.NoError(t, err)

	s := httptest.NewServer(h)
	defer s.Close()

	client, err := api.NewClient(s.URL, api.WithClient(s.Client()))
	require.NoError(t, err)

	t.Run("OptionalAnyParams", func(t *testing.T) {
		a := require.New(t)

		// Test with all parameters provided
		resp, err := client.AnyParams(ctx, api.AnyParamsParams{
			PathParam:    "test-path",
			QueryParam:   "test-query",
			XHeaderParam: "test-header",
			CookieParam:  "test-cookie",
		})
		a.NoError(err)
		a.Equal("test-path", resp.Echo.PathParam)
		a.True(resp.Echo.QueryParam.IsSet())
		a.Equal("test-query", resp.Echo.QueryParam.Value)
		a.True(resp.Echo.HeaderParam.IsSet())
		a.Equal("test-header", resp.Echo.HeaderParam.Value)
		a.True(resp.Echo.CookieParam.IsSet())
		a.Equal("test-cookie", resp.Echo.CookieParam.Value)

		// Test with only required parameter (path)
		// When optional any parameters are not provided, they remain nil
		resp, err = client.AnyParams(ctx, api.AnyParamsParams{
			PathParam: "only-path",
		})
		a.NoError(err)
		a.Equal("only-path", resp.Echo.PathParam)
		// Server doesn't set optional fields when they're nil
		a.False(resp.Echo.QueryParam.IsSet())
		a.False(resp.Echo.HeaderParam.IsSet())
		a.False(resp.Echo.CookieParam.IsSet())

		// Test with numeric-looking strings (should remain strings)
		resp, err = client.AnyParams(ctx, api.AnyParamsParams{
			PathParam:  "123",
			QueryParam: "456",
		})
		a.NoError(err)
		a.Equal("123", resp.Echo.PathParam)
		a.True(resp.Echo.QueryParam.IsSet())
		a.Equal("456", resp.Echo.QueryParam.Value)

		// Test with special characters
		resp, err = client.AnyParams(ctx, api.AnyParamsParams{
			PathParam:  "test%20value",
			QueryParam: "foo=bar&baz=qux",
		})
		a.NoError(err)
		a.Equal("test%20value", resp.Echo.PathParam)
		a.True(resp.Echo.QueryParam.IsSet())
		a.Equal("foo=bar&baz=qux", resp.Echo.QueryParam.Value)
	})

	t.Run("RequiredAnyParams", func(t *testing.T) {
		a := require.New(t)

		// Test with all required parameters
		resp, err := client.AnyParamsRequired(ctx, api.AnyParamsRequiredParams{
			PathParam:    "path-value",
			QueryParam:   "query-value",
			XHeaderParam: "header-value",
		})
		a.NoError(err)
		a.True(resp.Received)

		// Test that different string values work
		resp, err = client.AnyParamsRequired(ctx, api.AnyParamsRequiredParams{
			PathParam:    "true",
			QueryParam:   "false",
			XHeaderParam: "null",
		})
		a.NoError(err)
		a.True(resp.Received)
	})

	t.Run("ArrayOfAny", func(t *testing.T) {
		a := require.New(t)

		// Test with no parameter provided (nil array)
		resp, err := client.AnyArrayParam(ctx, api.AnyArrayParamParams{})
		a.NoError(err)
		// When array parameter is not provided, count is 0
		a.True(resp.Count.IsSet())
		a.Equal(0, resp.Count.Value)

		// Test with single item
		resp, err = client.AnyArrayParam(ctx, api.AnyArrayParamParams{
			Items: []any{"first"},
		})
		a.NoError(err)
		a.True(resp.Count.IsSet())
		a.Equal(1, resp.Count.Value)

		// Test with multiple items
		resp, err = client.AnyArrayParam(ctx, api.AnyArrayParamParams{
			Items: []any{"one", "two", "three"},
		})
		a.NoError(err)
		a.True(resp.Count.IsSet())
		a.Equal(3, resp.Count.Value)

		// Test with numeric-looking strings
		resp, err = client.AnyArrayParam(ctx, api.AnyArrayParamParams{
			Items: []any{"1", "2", "3"},
		})
		a.NoError(err)
		a.True(resp.Count.IsSet())
		a.Equal(3, resp.Count.Value)
	})
}
