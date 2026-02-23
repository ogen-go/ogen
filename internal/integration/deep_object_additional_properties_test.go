package integration

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_deep_object_additional_properties"
)

type testDeepObjectAdditionalProperties struct {
	api.UnimplementedHandler
	params api.QueryWithAdditionalPropertiesParams
}

func (s *testDeepObjectAdditionalProperties) QueryWithAdditionalProperties(ctx context.Context, params api.QueryWithAdditionalPropertiesParams) (api.QueryWithAdditionalPropertiesOK, error) {
	s.params = params
	obj, ok := params.Object.Get()
	if !ok {
		return api.QueryWithAdditionalPropertiesOK{}, nil
	}
	return api.QueryWithAdditionalPropertiesOK(obj), nil
}

func TestDeepObjectAdditionalProperties(t *testing.T) {
	handler := &testDeepObjectAdditionalProperties{}
	srv, err := api.NewServer(handler)
	require.NoError(t, err)

	s := httptest.NewServer(srv)
	defer s.Close()

	client, err := api.NewClient(s.URL)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("WithValues", func(t *testing.T) {
		input := api.QueryWithAdditionalPropertiesObject{
			"foo": "bar",
			"baz": "qux",
		}
		resp, err := client.QueryWithAdditionalProperties(ctx, api.QueryWithAdditionalPropertiesParams{
			Object: api.NewOptQueryWithAdditionalPropertiesObject(input),
		})
		require.NoError(t, err)
		require.Equal(t, "bar", map[string]string(resp)["foo"])
		require.Equal(t, "qux", map[string]string(resp)["baz"])
	})

	t.Run("WithoutValues", func(t *testing.T) {
		resp, err := client.QueryWithAdditionalProperties(ctx, api.QueryWithAdditionalPropertiesParams{})
		require.NoError(t, err)
		require.Empty(t, resp)
	})
}
