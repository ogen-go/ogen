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
		r = strings.NewReader(req.Name)
	case *api.AllRequestBodiesOptionalReqApplicationOctetStream:
		r = req
	case *api.AllRequestBodiesOptionalApplicationXWwwFormUrlencoded:
		r = strings.NewReader(req.Name)
	case *api.AllRequestBodiesOptionalMultipartFormData:
		r = strings.NewReader(req.Name)
	case *api.AllRequestBodiesOptionalReqTextPlain:
		r = req
	case *api.AllRequestBodiesOptionalReqEmptyBody:
		r = strings.NewReader("<empty body>")
	default:
		panic(fmt.Sprintf("unknown request type: %T", req))
	}

	return api.AllRequestBodiesOptionalOK{
		Data: r,
	}, nil
}

func (t testHTTPRequests) MaskContentType(ctx context.Context, req api.MaskContentTypeReqWithContentType) (api.MaskResponse, error) {
	var s strings.Builder
	if _, err := io.Copy(&s, req.Content); err != nil {
		return api.MaskResponse{}, err
	}
	return api.MaskResponse{
		ContentType: req.ContentType,
		Content:     s.String(),
	}, nil
}

func (t testHTTPRequests) MaskContentTypeOptional(ctx context.Context, req api.MaskContentTypeOptionalReqWithContentType) (api.MaskResponse, error) {
	var s strings.Builder
	if _, err := io.Copy(&s, req.Content); err != nil {
		return api.MaskResponse{}, err
	}
	return api.MaskResponse{
		ContentType: req.ContentType,
		Content:     s.String(),
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
	t.Run("AllRequestBodiesOptional", func(t *testing.T) {
		reqs := []api.AllRequestBodiesOptionalReq{
			&api.AllRequestBodiesOptionalApplicationJSON{
				Name: testData,
			},
			&api.AllRequestBodiesOptionalReqApplicationOctetStream{
				Data: strings.NewReader(testData),
			},
			&api.AllRequestBodiesOptionalApplicationXWwwFormUrlencoded{
				Name: testData,
			},
			&api.AllRequestBodiesOptionalMultipartFormData{
				Name: testData,
			},
			&api.AllRequestBodiesOptionalReqTextPlain{
				Data: strings.NewReader(testData),
			},
		}

		a := require.New(t)
		for _, req := range reqs {
			resp, err := client.AllRequestBodiesOptional(ctx, req)
			a.NoError(err)

			data, err := io.ReadAll(resp.Data)
			a.NoError(err)
			a.Equal(testData, string(data))
		}

		// Check that empty body is handled correctly.
		resp, err := client.AllRequestBodiesOptional(ctx, &api.AllRequestBodiesOptionalReqEmptyBody{})
		a.NoError(err)

		data, err := io.ReadAll(resp.Data)
		a.NoError(err)
		a.Equal("<empty body>", string(data))
	})
	t.Run("MaskContentType", func(t *testing.T) {
		a := require.New(t)

		_, err := client.MaskContentType(ctx, api.MaskContentTypeReqWithContentType{
			ContentType: "invalidCT",
			Content: api.MaskContentTypeReq{
				Data: strings.NewReader(testData),
			},
		})
		a.EqualError(err, `encode request: "invalidCT" does not match mask "application/*"`)

		resp, err := client.MaskContentType(ctx, api.MaskContentTypeReqWithContentType{
			ContentType: "application/json",
			Content: api.MaskContentTypeReq{
				Data: strings.NewReader(testData),
			},
		})
		a.NoError(err)
		a.Equal("application/json", resp.ContentType)
		a.Equal(testData, resp.Content)
	})
	t.Run("MaskContentTypeOptional", func(t *testing.T) {
		a := require.New(t)

		_, err := client.MaskContentTypeOptional(ctx, api.MaskContentTypeOptionalReqWithContentType{
			ContentType: "invalidCT",
			Content: api.MaskContentTypeOptionalReq{
				Data: strings.NewReader(testData),
			},
		})
		a.EqualError(err, `encode request: "invalidCT" does not match mask "application/*"`)

		resp, err := client.MaskContentTypeOptional(ctx, api.MaskContentTypeOptionalReqWithContentType{
			ContentType: "application/json",
			Content: api.MaskContentTypeOptionalReq{
				Data: strings.NewReader(testData),
			},
		})
		a.NoError(err)
		a.Equal("application/json", resp.ContentType)
		a.Equal(testData, resp.Content)
	})
}
