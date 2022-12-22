package integration

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/internal/integration/customformats/phonetype"
	"github.com/ogen-go/ogen/internal/integration/customformats/rgbatype"
	"github.com/ogen-go/ogen/internal/integration/test_customformats"
)

type testCustomFormats struct{}

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
	a := require.New(t)
	ctx := context.Background()

	srv, err := api.NewServer(testCustomFormats{})
	a.NoError(err)

	s := httptest.NewServer(srv)
	defer s.Close()

	client, err := api.NewClient(s.URL, api.WithClient(s.Client()))
	a.NoError(err)

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
}
