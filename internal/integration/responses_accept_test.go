package integration_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/http"
	api "github.com/ogen-go/ogen/internal/integration/test_http_responses_accept"
)

var (
	errNotAcceptable = errors.New("not acceptable")
)

type testHTTPResponsesAccept struct{}

func (t testHTTPResponsesAccept) MultipleContentTypesWithParameters(ctx context.Context, params api.MultipleContentTypesWithParametersParams) (api.MultipleContentTypesWithParametersRes, error) {
	if params.Accept.MatchesContentType(api.MediaTypeApplicationOctetStream) {
		return &api.MultipleContentTypesWithParametersOKApplicationOctetStream{
			Data: bytes.NewBufferString("byte content with parameter " + params.Q),
		}, nil
	} else if params.Accept.MatchesContentType(api.MediaTypeApplicationJSON) {
		return &api.MultipleContentTypesWithParametersOKApplicationJSON{
			Data: "json data with parameter " + params.Q,
		}, nil
	} else {
		return nil, errNotAcceptable
	}
}

func (t testHTTPResponsesAccept) MultipleContentTypesWithoutParameters(ctx context.Context, params api.MultipleContentTypesWithoutParametersParams) (api.MultipleContentTypesWithoutParametersRes, error) {
	if params.Accept.MatchesContentType(api.MediaTypeApplicationOctetStream) {
		return &api.MultipleContentTypesWithoutParametersOKApplicationOctetStream{
			Data: bytes.NewBufferString("byte content"),
		}, nil
	} else if params.Accept.MatchesContentType(api.MediaTypeApplicationJSON) {
		return &api.MultipleContentTypesWithoutParametersOKApplicationJSON{
			Data: "json data",
		}, nil
	} else {
		return nil, errNotAcceptable
	}
}

func testResponsesAcceptInit(t *testing.T, h testHTTPResponsesAccept) (*require.Assertions, *api.Client) {
	a := require.New(t)

	srv, err := api.NewServer(h)
	a.NoError(err)

	s := httptest.NewServer(srv)
	t.Cleanup(func() {
		s.Close()
	})

	client, err := api.NewClient(s.URL, api.WithClient(s.Client()))
	a.NoError(err)

	return a, client
}

func TestResponsesAcceptMultipleContentTypes(t *testing.T) {
	create := func(t *testing.T) (context.Context, *require.Assertions, *api.Client) {
		a, client := testResponsesAcceptInit(t, testHTTPResponsesAccept{})
		return context.Background(), a, client
	}

	ctx, a, client := create(t)

	t.Run("Without Parameters", func(t *testing.T) {
		t.Run("byte data", func(t *testing.T) {

			r, err := client.MultipleContentTypesWithoutParameters(ctx, api.MultipleContentTypesWithoutParametersParams{
				Accept: http.AcceptHeaderNew(api.MediaTypeApplicationOctetStream),
			})
			a.NoError(err)

			a.IsType(&api.MultipleContentTypesWithoutParametersOKApplicationOctetStream{}, r)
			res := r.(*api.MultipleContentTypesWithoutParametersOKApplicationOctetStream)

			data, err := io.ReadAll(res.Data)
			a.NoError(err)
			a.Equal([]byte("byte content"), data)
		})
		t.Run("json data", func(t *testing.T) {

			r, err := client.MultipleContentTypesWithoutParameters(ctx, api.MultipleContentTypesWithoutParametersParams{
				Accept: http.AcceptHeaderNew(api.MediaTypeApplicationJSON),
			})
			a.NoError(err)

			a.IsType(&api.MultipleContentTypesWithoutParametersOKApplicationJSON{}, r)
			res := r.(*api.MultipleContentTypesWithoutParametersOKApplicationJSON)

			a.Equal("json data", res.Data)
		})
	})

	t.Run("With Parameters", func(t *testing.T) {
		t.Run("byte data", func(t *testing.T) {

			r, err := client.MultipleContentTypesWithParameters(ctx, api.MultipleContentTypesWithParametersParams{
				Accept: http.AcceptHeaderNew(api.MediaTypeApplicationOctetStream),
				Q:      "from unit test",
			})
			a.NoError(err)

			a.IsType(&api.MultipleContentTypesWithParametersOKApplicationOctetStream{}, r)
			res := r.(*api.MultipleContentTypesWithParametersOKApplicationOctetStream)

			data, err := io.ReadAll(res.Data)
			a.NoError(err)
			a.Equal([]byte("byte content with parameter from unit test"), data)
		})
		t.Run("json data", func(t *testing.T) {

			r, err := client.MultipleContentTypesWithParameters(ctx, api.MultipleContentTypesWithParametersParams{
				Accept: http.AcceptHeaderNew(api.MediaTypeApplicationJSON),
				Q:      "from unit test",
			})
			a.NoError(err)

			a.IsType(&api.MultipleContentTypesWithParametersOKApplicationJSON{}, r)
			res := r.(*api.MultipleContentTypesWithParametersOKApplicationJSON)

			a.Equal("json data with parameter from unit test", res.Data)
		})
	})
}
