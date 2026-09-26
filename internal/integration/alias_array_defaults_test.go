package integration

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_alias_array_defaults"
)

type aliasArrayDefaultsHandler struct{}

func (aliasArrayDefaultsHandler) CreateItems(_ context.Context, req *api.CreateItemsRequest) (*api.CreateItemsRequest, error) {
	return req, nil
}

// TestAliasArrayDefaults verifies that defaults of properties referencing a
// named array schema ($ref, optionally wrapped in anyOf with null) are applied.
func TestAliasArrayDefaults(t *testing.T) {
	expected := api.CreateItemsRequest{
		Tags:         api.TagList{"a", "b"},
		NullableTags: api.NewOptNilTagList(api.TagList{"c"}),
		Colors:       api.ColorList{api.ColorBlue},
		Items:        api.ItemList{{Name: api.NewOptString("x")}},
		TagLists:     []api.TagList{{"d"}, {"e", "f"}},
	}

	t.Run("Decode", func(t *testing.T) {
		var got api.CreateItemsRequest
		require.NoError(t, got.Decode(jx.DecodeStr(`{}`)))
		require.Equal(t, expected, got)
	})

	t.Run("ExplicitValue", func(t *testing.T) {
		var got api.CreateItemsRequest
		require.NoError(t, got.Decode(jx.DecodeStr(`{"tags":["z"],"nullableTags":null}`)))
		require.Equal(t, api.TagList{"z"}, got.Tags)
		require.True(t, got.NullableTags.IsNull())
		require.Equal(t, expected.Colors, got.Colors)
	})

	t.Run("Server", func(t *testing.T) {
		srv, err := api.NewServer(aliasArrayDefaultsHandler{})
		require.NoError(t, err)

		s := httptest.NewServer(srv)
		t.Cleanup(s.Close)

		client, err := api.NewClient(s.URL, api.WithClient(s.Client()))
		require.NoError(t, err)

		// Zero-value request omits all fields, so the server applies defaults.
		got, err := client.CreateItems(context.Background(), &api.CreateItemsRequest{})
		require.NoError(t, err)
		require.Equal(t, expected, *got)
	})
}
