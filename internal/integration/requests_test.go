package integration_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_http_requests"
)

type testHTTPRequests struct{}

func (t testHTTPRequests) Base64Request(ctx context.Context, req api.Base64RequestReq) (api.Base64RequestOK, error) {
	return api.Base64RequestOK{
		Data: req,
	}, nil
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
	case *api.SimpleObjectMultipart:
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
	case *api.SimpleObjectMultipart:
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

func (t testHTTPRequests) MaskContentType(ctx context.Context, req *api.MaskContentTypeReqWithContentType) (*api.MaskResponse, error) {
	var s strings.Builder
	if _, err := io.Copy(&s, req.Content); err != nil {
		return nil, err
	}
	return &api.MaskResponse{
		ContentType: req.ContentType,
		Content:     s.String(),
	}, nil
}

func (t testHTTPRequests) MaskContentTypeOptional(ctx context.Context, req *api.MaskContentTypeOptionalReqWithContentType) (*api.MaskResponse, error) {
	var s strings.Builder
	if _, err := io.Copy(&s, req.Content); err != nil {
		return nil, err
	}
	return &api.MaskResponse{
		ContentType: req.ContentType,
		Content:     s.String(),
	}, nil
}

func (t testHTTPRequests) StreamJSON(ctx context.Context, req []float64) (v float64, _ error) {
	for _, f := range req {
		v += f
	}
	return v, nil
}

func TestRequests(t *testing.T) {
	ctx := context.Background()

	testData := "bababoi"
	h, err := api.NewServer(testHTTPRequests{})
	require.NoError(t, err)

	s := httptest.NewServer(h)
	defer s.Close()

	client, err := api.NewClient(s.URL, api.WithClient(s.Client()))
	require.NoError(t, err)

	t.Run("AllRequestBodies", func(t *testing.T) {
		reqs := []api.AllRequestBodiesReq{
			&api.AllRequestBodiesApplicationJSON{
				SimpleObject: api.SimpleObject{
					Name: testData,
				},
			},
			&api.AllRequestBodiesReqApplicationOctetStream{
				Data: strings.NewReader(testData),
			},
			&api.AllRequestBodiesApplicationXWwwFormUrlencoded{
				SimpleObject: api.SimpleObject{
					Name: testData,
				},
			},
			&api.SimpleObjectMultipart{
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
				SimpleObject: api.SimpleObject{
					Name: testData,
				},
			},
			&api.AllRequestBodiesOptionalReqApplicationOctetStream{
				Data: strings.NewReader(testData),
			},
			&api.AllRequestBodiesOptionalApplicationXWwwFormUrlencoded{
				SimpleObject: api.SimpleObject{
					Name: testData,
				},
			},
			&api.SimpleObjectMultipart{
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

		_, err := client.MaskContentType(ctx, &api.MaskContentTypeReqWithContentType{
			ContentType: "invalidCT",
			Content: api.MaskContentTypeReq{
				Data: strings.NewReader(testData),
			},
		})
		a.EqualError(err, `encode request: "invalidCT" does not match mask "application/*"`)

		resp, err := client.MaskContentType(ctx, &api.MaskContentTypeReqWithContentType{
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

		_, err := client.MaskContentTypeOptional(ctx, &api.MaskContentTypeOptionalReqWithContentType{
			ContentType: "invalidCT",
			Content: api.MaskContentTypeOptionalReq{
				Data: strings.NewReader(testData),
			},
		})
		a.EqualError(err, `encode request: "invalidCT" does not match mask "application/*"`)

		resp, err := client.MaskContentTypeOptional(ctx, &api.MaskContentTypeOptionalReqWithContentType{
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

func TestRequestBase64(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()

	testData := "bababoi"
	srv, err := api.NewServer(testHTTPRequests{})
	a.NoError(err)

	s := httptest.NewServer(srv)
	defer s.Close()

	client, err := api.NewClient(s.URL, api.WithClient(s.Client()))
	a.NoError(err)

	resp, err := client.Base64Request(ctx, api.Base64RequestReq{
		Data: strings.NewReader(testData),
	})
	a.NoError(err)

	var sb strings.Builder
	_, err = io.Copy(&sb, resp.Data)
	a.NoError(err)
	a.Equal(testData, sb.String())

	{
		encoded := base64.StdEncoding.EncodeToString([]byte(testData))
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			s.URL+"/base64Request",
			strings.NewReader(encoded),
		)
		a.NoError(err)
		req.Header.Set("Content-Type", "text/plain")

		resp, err := s.Client().Do(req)
		a.NoError(err)
		defer resp.Body.Close()

		a.Equal(http.StatusOK, resp.StatusCode)
		a.Equal("text/plain; charset=utf-8", resp.Header.Get("Content-Type"))

		var sb strings.Builder
		_, err = io.Copy(&sb, base64.NewDecoder(base64.StdEncoding, resp.Body))
		a.NoError(err)
		a.Equal(testData, sb.String())
	}
}

func TestRequestJSONTrailingData(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()

	testData := "bababoi"
	srv, err := api.NewServer(testHTTPRequests{},
		api.WithErrorHandler(func(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, err.Error())
		}),
	)
	a.NoError(err)

	s := httptest.NewServer(srv)
	defer s.Close()

	send := func(reqBody string) (code int, response string) {
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			s.URL+"/allRequestBodies",
			strings.NewReader(reqBody),
		)
		a.NoError(err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.Client().Do(req)
		a.NoError(err)
		defer func() {
			_ = resp.Body.Close()
		}()

		var sb strings.Builder
		_, err = io.Copy(&sb, resp.Body)
		a.NoError(err)
		return resp.StatusCode, sb.String()
	}

	code, resp := send(fmt.Sprintf(`{"name":%q}{"name":"trailing"}`, testData))
	a.Equal(http.StatusBadRequest, code)
	a.Contains(resp, ": unexpected trailing data")

	// Trailing newlines are ok.
	code, resp = send(fmt.Sprintf("{\"name\":%q}\n\n", testData))
	a.Equal(http.StatusOK, code)
	a.Equal(testData, resp)
}

func TestServerURLOverride(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()

	testData := "bababoi"
	srv, err := api.NewServer(testHTTPRequests{})
	a.NoError(err)

	s := httptest.NewServer(srv)
	defer s.Close()

	client, err := api.NewClient("https://example.com", api.WithClient(s.Client()))
	a.NoError(err)

	override, err := url.Parse(s.URL)
	a.NoError(err)

	// Send request to the server, not to the example.com.
	resp, err := client.MaskContentType(api.WithServerURL(ctx, override), &api.MaskContentTypeReqWithContentType{
		ContentType: "application/json",
		Content: api.MaskContentTypeReq{
			Data: strings.NewReader(testData),
		},
	})
	a.NoError(err)
	a.Equal("application/json", resp.ContentType)
	a.Equal(testData, resp.Content)
}

func TestServerURLTrimSlashes(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()

	testData := "bababoi"
	srv, err := api.NewServer(testHTTPRequests{})
	a.NoError(err)

	s := httptest.NewServer(srv)
	defer s.Close()

	hclient := s.Client()
	for _, u := range []string{
		s.URL,
		s.URL + "/",
		s.URL + "//",
	} {
		u := u
		t.Logf("Server: %q", u)

		client, err := api.NewClient(u, api.WithClient(hclient))
		a.NoError(err)

		resp, err := client.MaskContentType(ctx, &api.MaskContentTypeReqWithContentType{
			ContentType: "application/json",
			Content: api.MaskContentTypeReq{
				Data: strings.NewReader(testData),
			},
		})
		a.NoError(err)
		a.Equal("application/json", resp.ContentType)
		a.Equal(testData, resp.Content)
	}
}

func TestRequestJSONStreaming(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()

	srv, err := api.NewServer(testHTTPRequests{})
	a.NoError(err)

	s := httptest.NewServer(srv)
	defer s.Close()

	client, err := api.NewClient(s.URL, api.WithClient(s.Client()))
	a.NoError(err)

	r, err := client.StreamJSON(ctx, []float64{1, 2, 3})
	a.NoError(err)
	a.Equal(int(6), int(r))
}
