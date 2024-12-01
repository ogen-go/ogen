package integration

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/sample_err"
)

type sampleErrServer struct{}

func (s sampleErrServer) DataCreate(ctx context.Context, req api.OptData) (*api.Data, error) {
	panic("implement me")
}

func (s sampleErrServer) DataGet(ctx context.Context) (*api.Data, error) {
	return nil, &api.ErrorStatusCode{
		StatusCode: 500,
		Response: api.Error{
			Code:    -200,
			Message: "Ok (false)",
		},
	}
}

func (s sampleErrServer) NewError(ctx context.Context, err error) *api.ErrorStatusCode {
	panic("should not be called")
}

func TestConvenientErrors(t *testing.T) {
	h, err := api.NewServer(&sampleErrServer{})
	require.NoError(t, err)

	s := httptest.NewServer(h)
	defer s.Close()

	client, err := api.NewClient(s.URL)
	require.NoError(t, err)
	ctx := context.Background()

	_, getErr := client.DataGet(ctx)
	var statusErr *api.ErrorStatusCode
	require.ErrorAs(t, getErr, &statusErr)
	require.Equal(t, "code 500: {Code:-200 Message:Ok (false)}", statusErr.Error())
}
