package integration

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/internal/integration/customformats/phonetype"
	"github.com/ogen-go/ogen/internal/integration/customformats/rgbatype"
	api "github.com/ogen-go/ogen/internal/integration/test_customformats"
)

type testCustomFormats struct{}

func (t testCustomFormats) EventPost(ctx context.Context, req any) (any, error) {
	if req == nil {
		return nil, errors.New("empty request")
	}
	return req, nil
}

func (t testCustomFormats) PhoneGet(ctx context.Context, req *api.User, params api.PhoneGetParams) (*api.User, error) {
	req.HomePhone.SetTo(params.Phone)
	if v, ok := params.Color.Get(); ok {
		req.BackgroundColor.SetTo(v)
	}
	if v, ok := params.Hex.Get(); ok {
		req.HexColor.SetTo(v)
	}
	return req, nil
}

func TestCustomFormats(t *testing.T) {
	ctx := context.Background()

	srv, err := api.NewServer(testCustomFormats{})
	require.NoError(t, err)

	s := httptest.NewServer(srv)
	defer s.Close()

	client, err := api.NewClient(s.URL, api.WithClient(s.Client()))
	require.NoError(t, err)

	t.Run("EventPost", func(t *testing.T) {
		a := require.New(t)

		for _, val := range []any{
			true,
			float64(42),
			"string",
			[]any{float64(1), float64(2), float64(3)},
			map[string]any{
				"key": []any{"value", "value2"},
			},
		} {
			result, err := client.EventPost(ctx, val)
			a.NoError(err)
			a.Equal(val, result)
		}
	})
	t.Run("Phone", func(t *testing.T) {
		a := require.New(t)

		var (
			homePhone       = phonetype.Phone("+1234567890")
			backgroundColor = rgbatype.RGBA{R: 255, G: 0, B: 0, A: 255}
			hex             = int64(100)

			u = &api.User{
				ID:           10,
				Phone:        "+1234567890",
				ProfileColor: rgbatype.RGBA{R: 0, G: 0, B: 0, A: 255},
			}
		)

		u2, err := client.PhoneGet(ctx, u, api.PhoneGetParams{
			Phone: homePhone,
			Color: api.NewOptRgba(backgroundColor),
			Hex:   api.NewOptHex(hex),
		})
		a.NoError(err)

		a.Equal(u.ID, u2.ID)
		a.Equal(u.Phone, u2.Phone)
		a.Equal(u.ProfileColor, u2.ProfileColor)
		a.Equal(homePhone, u2.HomePhone.Or(""))
		a.Equal(backgroundColor, u2.BackgroundColor.Or(rgbatype.RGBA{}))
		a.Equal(hex, u2.HexColor.Or(0))
	})
}
