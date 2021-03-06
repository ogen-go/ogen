package internal_test

import (
	"context"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/test_http_requests"
)

type testHTTPRequests struct {
}

func (t testHTTPRequests) AllRequestBodies(
	_ context.Context,
	req api.AllRequestBodiesReq,
) (api.AllRequestBodiesOK, error) {
	var r io.Reader

	switch req := req.(type) {
	case *api.AllRequestBodiesApplicationJSON:
		r = strings.NewReader(req.Name)
	case *api.AllRequestBodiesReqApplicationOctetStream:
		r = req
	case *api.AllRequestBodiesApplicationXWwwFormUrlencoded:
		r = strings.NewReader(req.Name)
	case *api.AllRequestBodiesMultipartFormData:
		r = strings.NewReader(req.Name)
	case *api.AllRequestBodiesReqTextPlain:
		r = req
	default:
		panic(fmt.Sprintf("unknown request type: %T", req))
	}

	return api.AllRequestBodiesOK{
		Data: r,
	}, nil
}

func (t testHTTPRequests) AllRequestBodiesOptional(
	_ context.Context,
	req api.AllRequestBodiesOptionalReq,
) (api.AllRequestBodiesOptionalOK, error) {
	var r io.Reader

	switch req := req.(type) {
	case *api.AllRequestBodiesOptionalApplicationJSON:
		if val, ok := api.OptSimpleObject(*req).Get(); ok {
			r = strings.NewReader(val.Name)
		}
	case *api.AllRequestBodiesOptionalReqApplicationOctetStream:
		r = req
	case *api.AllRequestBodiesOptionalApplicationXWwwFormUrlencoded:
		if val, ok := api.OptSimpleObject(*req).Get(); ok {
			r = strings.NewReader(val.Name)
		}
	case *api.AllRequestBodiesOptionalMultipartFormData:
		if val, ok := api.OptSimpleObject(*req).Get(); ok {
			r = strings.NewReader(val.Name)
		}
	case *api.AllRequestBodiesOptionalReqTextPlain:
		r = req
	default:
		panic(fmt.Sprintf("unknown request type: %T", req))
	}

	return api.AllRequestBodiesOptionalOK{
		Data: r,
	}, nil
}

func TestRequests(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()

	testData := "bababoi"
	srv, err := api.NewServer(testHTTPRequests{})
	a.NoError(err)

	s := httptest.NewServer(srv)
	defer s.Close()

	client, err := api.NewClient(s.URL, api.WithClient(s.Client()))
	a.NoError(err)

	t.Run("AllRequestBodies", func(t *testing.T) {
		reqs := []api.AllRequestBodiesReq{
			&api.AllRequestBodiesApplicationJSON{
				Name: testData,
			},
			&api.AllRequestBodiesReqApplicationOctetStream{
				Data: strings.NewReader(testData),
			},
			&api.AllRequestBodiesApplicationXWwwFormUrlencoded{
				Name: testData,
			},
			&api.AllRequestBodiesMultipartFormData{
				Name: testData,
			},
			&api.AllRequestBodiesReqTextPlain{
				Data: strings.NewReader(testData),
			},
		}

		a := require.New(t)
		for _, req := range reqs {
			resp, err := client.AllRequestBodies(ctx, req)
			a.NoError(err)

			data, err := io.ReadAll(resp.Data)
			a.NoError(err)
			a.Equal(testData, string(data))
		}
	})
}
