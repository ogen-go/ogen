package integration

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http/httptest"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_parameters"
)

type testParameters struct {
}

func (s *testParameters) ObjectQueryParameter(ctx context.Context, params api.ObjectQueryParameterParams) (*api.ObjectQueryParameterOK, error) {
	if param, ok := params.FormObject.Get(); ok {
		return &api.ObjectQueryParameterOK{
			Style:  "form",
			Min:    param.Min,
			Max:    param.Max,
			Filter: param.Filter,
		}, nil
	}
	if param, ok := params.DeepObject.Get(); ok {
		return &api.ObjectQueryParameterOK{
			Style:  "deepObject",
			Min:    param.Min,
			Max:    param.Max,
			Filter: param.Filter,
		}, nil
	}
	return &api.ObjectQueryParameterOK{}, errors.New("invalid input")
}

func (s *testParameters) ContentParameters(ctx context.Context, params api.ContentParametersParams) (*api.ContentParameters, error) {
	return &api.ContentParameters{
		Query:  params.Query,
		Path:   params.Path,
		Header: params.XHeader,
		Cookie: params.Cookie,
	}, nil
}

func (s *testParameters) HeaderParameter(ctx context.Context, params api.HeaderParameterParams) (*api.Hash, error) {
	h := sha256.Sum256([]byte(params.XAuthToken))
	return &api.Hash{
		Raw: h[:],
		Hex: hex.EncodeToString(h[:]),
	}, nil
}

func (s *testParameters) CookieParameter(ctx context.Context, params api.CookieParameterParams) (*api.Hash, error) {
	h := sha256.Sum256([]byte(params.Value))
	return &api.Hash{
		Raw: h[:],
		Hex: hex.EncodeToString(h[:]),
	}, nil
}

func (s *testParameters) ComplicatedParameterNameGet(ctx context.Context, params api.ComplicatedParameterNameGetParams) error {
	panic("implement me")
}

func (s *testParameters) SameName(ctx context.Context, params api.SameNameParams) error {
	panic("implement me")
}

func TestParameters(t *testing.T) {
	ctx := context.Background()

	h, err := api.NewServer(&testParameters{})
	require.NoError(t, err)

	s := httptest.NewServer(h)
	defer s.Close()

	client, err := api.NewClient(s.URL, api.WithClient(s.Client()))
	require.NoError(t, err)

	t.Run("ObjectQueryParameter", func(t *testing.T) {
		const (
			min    = 1
			max    = 5
			filter = "abc"
		)

		t.Run("formStyle", func(t *testing.T) {
			resp, err := client.ObjectQueryParameter(ctx, api.ObjectQueryParameterParams{
				FormObject: api.NewOptObjectQueryParameterFormObject(api.ObjectQueryParameterFormObject{
					Min:    min,
					Max:    max,
					Filter: filter,
				}),
			})
			require.NoError(t, err)
			require.Equal(t, resp.Style, "form")
			require.Equal(t, resp.Min, min)
			require.Equal(t, resp.Max, max)
			require.Equal(t, resp.Filter, filter)
		})
		t.Run("deepObjectStyle", func(t *testing.T) {
			resp, err := client.ObjectQueryParameter(ctx, api.ObjectQueryParameterParams{
				DeepObject: api.NewOptObjectQueryParameterDeepObject(api.ObjectQueryParameterDeepObject{
					Min:    min,
					Max:    max,
					Filter: filter,
				}),
			})
			require.NoError(t, err)
			require.Equal(t, resp.Style, "deepObject")
			require.Equal(t, resp.Min, min)
			require.Equal(t, resp.Max, max)
			require.Equal(t, resp.Filter, filter)
		})
	})

	t.Run("HeaderParameter", func(t *testing.T) {
		h, err := client.HeaderParameter(ctx, api.HeaderParameterParams{XAuthToken: "hello, world"})
		require.NoError(t, err)
		assert.NotEmpty(t, h.Raw)
		assert.Equal(t, hex.EncodeToString(h.Raw), h.Hex)
		assert.Equal(t, "09ca7e4eaa6e8ae9c7d261167129184883644d07dfba7cbfbc4c8a2e08360d5b", h.Hex)
	})
	t.Run("CookieParameter", func(t *testing.T) {
		h, err := client.CookieParameter(ctx, api.CookieParameterParams{Value: "hello, world"})
		require.NoError(t, err)
		assert.NotEmpty(t, h.Raw)
		assert.Equal(t, hex.EncodeToString(h.Raw), h.Hex)
		assert.Equal(t, "09ca7e4eaa6e8ae9c7d261167129184883644d07dfba7cbfbc4c8a2e08360d5b", h.Hex)
	})

	t.Run("ContentParameters", func(t *testing.T) {
		user := api.User{
			ID:       1,
			Username: "admin",
			Role:     api.UserRoleAdmin,
			Friends: []api.User{
				{
					ID:       2,
					Username: "alice",
					Role:     api.UserRoleUser,
				},
				{
					ID:       3,
					Username: "`\"';,./<>?[]{}\\|~!@#$%^&*()_+-=",
					Role:     api.UserRoleBot,
				},
			},
		}
		resp, err := client.ContentParameters(ctx, api.ContentParametersParams{
			Query:   user,
			Path:    user,
			XHeader: user,
			Cookie:  user,
		})
		require.NoError(t, err)
		assert.Equal(t, user, resp.Query)
		assert.Equal(t, user, resp.Path)
		assert.Equal(t, user, resp.Header)
		assert.Equal(t, user, resp.Cookie)
	})
}
