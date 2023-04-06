package integration_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_http_responses"
	"github.com/ogen-go/ogen/validate"
)

type testHTTPResponses struct {
	data []byte

	headerUser    api.User
	headerJSONRaw jx.Raw
}

func (t testHTTPResponses) AnyContentTypeBinaryStringSchema(ctx context.Context) (api.AnyContentTypeBinaryStringSchemaOK, error) {
	return api.AnyContentTypeBinaryStringSchemaOK{
		Data: bytes.NewReader(t.data),
	}, nil
}

func (t testHTTPResponses) AnyContentTypeBinaryStringSchemaDefault(ctx context.Context) (*api.AnyContentTypeBinaryStringSchemaDefaultDefStatusCode, error) {
	return &api.AnyContentTypeBinaryStringSchemaDefaultDefStatusCode{
		StatusCode: 200,
		Response: api.AnyContentTypeBinaryStringSchemaDefaultDef{
			Data: bytes.NewReader(t.data),
		},
	}, nil
}

func (t testHTTPResponses) MultipleGenericResponses(ctx context.Context) (api.MultipleGenericResponsesRes, error) {
	return &api.NilString{
		Null: true,
	}, nil
}

func (t testHTTPResponses) OctetStreamBinaryStringSchema(ctx context.Context) (api.OctetStreamBinaryStringSchemaOK, error) {
	return api.OctetStreamBinaryStringSchemaOK{
		Data: bytes.NewReader(t.data),
	}, nil
}

func (t testHTTPResponses) OctetStreamEmptySchema(ctx context.Context) (api.OctetStreamEmptySchemaOK, error) {
	return api.OctetStreamEmptySchemaOK{
		Data: bytes.NewReader(t.data),
	}, nil
}

func (t testHTTPResponses) TextPlainBinaryStringSchema(ctx context.Context) (api.TextPlainBinaryStringSchemaOK, error) {
	return api.TextPlainBinaryStringSchemaOK{
		Data: bytes.NewReader(t.data),
	}, nil
}

func (t testHTTPResponses) IntersectPatternCode(ctx context.Context, params api.IntersectPatternCodeParams) (api.IntersectPatternCodeRes, error) {
	if params.Code == 200 {
		var resp api.IntersectPatternCodeOKApplicationJSON = "200"
		return &resp, nil
	}
	return &api.IntersectPatternCode2XXStatusCode{
		StatusCode: params.Code,
		Response:   params.Code,
	}, nil
}

func (t testHTTPResponses) Combined(ctx context.Context, params api.CombinedParams) (api.CombinedRes, error) {
	switch params.Type {
	case api.CombinedType200:
		return &api.CombinedOK{
			Ok: "200",
		}, nil
	case api.CombinedType2XX:
		return &api.Combined2XXStatusCode{
			StatusCode: http.StatusAccepted,
			Response:   http.StatusAccepted,
		}, nil
	case api.CombinedType5XX:
		return &api.Combined5XXStatusCode{
			StatusCode: http.StatusInternalServerError,
			Response:   true,
		}, nil
	case api.CombinedTypeDefault:
		return &api.CombinedDefStatusCode{
			StatusCode: http.StatusNotFound,
			Response:   []string{"default"},
		}, nil
	default:
		panic(fmt.Sprintf("unknown type %q", params.Type))
	}
}

func (t testHTTPResponses) Headers200(ctx context.Context) (*api.Headers200OK, error) {
	return &api.Headers200OK{
		XTestHeader: "foo",
	}, nil
}

func (t testHTTPResponses) HeadersDefault(ctx context.Context) (*api.HeadersDefaultDef, error) {
	return &api.HeadersDefaultDef{
		XTestHeader: "202",
		StatusCode:  202,
	}, nil
}

func (t testHTTPResponses) HeadersPattern(ctx context.Context) (*api.HeadersPattern4XX, error) {
	return &api.HeadersPattern4XX{
		XTestHeader: "404",
		StatusCode:  404,
	}, nil
}

func (t testHTTPResponses) HeadersCombined(ctx context.Context, params api.HeadersCombinedParams) (api.HeadersCombinedRes, error) {
	switch params.Type {
	case api.HeadersCombinedType200:
		return &api.HeadersCombinedOK{
			XTestHeader: "200",
		}, nil
	case api.HeadersCombinedTypeDefault:
		return &api.HeadersCombinedDef{
			XTestHeader: "default",
			StatusCode:  202,
		}, nil
	case api.HeadersCombinedType4XX:
		return &api.HeadersCombined4XX{
			XTestHeader: "4XX",
			StatusCode:  404,
		}, nil
	default:
		panic(fmt.Sprintf("unknown type %q", params.Type))
	}
}

func (t testHTTPResponses) HeadersJSON(ctx context.Context) (*api.HeadersJSONOK, error) {
	return &api.HeadersJSONOK{
		XJSONCustomHeader: t.headerJSONRaw,
		XJSONHeader:       t.headerUser,
	}, nil
}

func (t testHTTPResponses) OptionalHeaders(ctx context.Context) (*api.OptionalHeadersOK, error) {
	return &api.OptionalHeadersOK{
		XOptional: api.OptString{},
		XRequired: "required",
	}, nil
}

func (t testHTTPResponses) StreamJSON(ctx context.Context, params api.StreamJSONParams) (api.StreamJSONRes, error) {
	n := make(api.QueryData, params.Count)
	for i := range n {
		n[i] = rand.NormFloat64()
	}
	return &n, nil
}

func testResponsesInit(t *testing.T, h testHTTPResponses) (*require.Assertions, *api.Client) {
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

func TestResponsesEncoding(t *testing.T) {
	testData := []byte("bababoi")
	create := func(t *testing.T) (context.Context, *require.Assertions, *api.Client) {
		a, client := testResponsesInit(t, testHTTPResponses{
			data: testData,
		})
		return context.Background(), a, client
	}

	t.Run("AnyContentTypeBinaryStringSchema", func(t *testing.T) {
		ctx, a, client := create(t)

		r, err := client.AnyContentTypeBinaryStringSchema(ctx)
		a.NoError(err)
		data, err := io.ReadAll(r.Data)
		a.NoError(err)
		a.Equal(testData, data)
	})
	t.Run("AnyContentTypeBinaryStringSchemaDefault", func(t *testing.T) {
		ctx, a, client := create(t)

		r, err := client.AnyContentTypeBinaryStringSchemaDefault(ctx)
		a.NoError(err)
		data, err := io.ReadAll(r.Response.Data)
		a.NoError(err)
		a.Equal(testData, data)
		a.Equal(200, r.StatusCode)
	})
	t.Run("MultipleGenericResponses", func(t *testing.T) {
		ctx, a, client := create(t)

		r, err := client.MultipleGenericResponses(ctx)
		a.NoError(err)
		a.Equal(&api.NilString{Null: true}, r)
	})
	t.Run("OctetStreamBinaryStringSchema", func(t *testing.T) {
		ctx, a, client := create(t)

		r, err := client.OctetStreamBinaryStringSchema(ctx)
		a.NoError(err)
		data, err := io.ReadAll(r.Data)
		a.NoError(err)
		a.Equal(testData, data)
	})
	t.Run("OctetStreamEmptySchema", func(t *testing.T) {
		ctx, a, client := create(t)

		r, err := client.OctetStreamEmptySchema(ctx)
		a.NoError(err)
		data, err := io.ReadAll(r.Data)
		a.NoError(err)
		a.Equal(testData, data)
	})
	t.Run("TextPlainBinaryStringSchema", func(t *testing.T) {
		ctx, a, client := create(t)

		r, err := client.TextPlainBinaryStringSchema(ctx)
		a.NoError(err)
		data, err := io.ReadAll(r.Data)
		a.NoError(err)
		a.Equal(testData, data)
	})
}

func TestResponsesHeaders(t *testing.T) {
	testData := []byte("bababoi")
	create := func(t *testing.T) (context.Context, *require.Assertions, *api.Client) {
		a, client := testResponsesInit(t, testHTTPResponses{
			data: testData,
		})
		return context.Background(), a, client
	}

	t.Run("Headers200", func(t *testing.T) {
		ctx, a, client := create(t)

		r, err := client.Headers200(ctx)
		a.NoError(err)
		a.Equal(r.XTestHeader, "foo")
	})
	t.Run("HeadersDefault", func(t *testing.T) {
		ctx, a, client := create(t)

		r, err := client.HeadersDefault(ctx)
		a.NoError(err)
		a.Equal(r.StatusCode, 202)
		a.Equal(r.XTestHeader, "202")
	})
	t.Run("HeadersPattern", func(t *testing.T) {
		ctx, a, client := create(t)

		r, err := client.HeadersPattern(ctx)
		a.NoError(err)
		a.Equal(r.StatusCode, 404)
		a.Equal(r.XTestHeader, "404")
	})
	t.Run("HeadersCombined", func(t *testing.T) {
		tests := []struct {
			Param    api.HeadersCombinedType
			Response api.HeadersCombinedRes
		}{
			{
				api.HeadersCombinedType200,
				&api.HeadersCombinedOK{XTestHeader: "200"},
			},
			{
				api.HeadersCombinedTypeDefault,
				&api.HeadersCombinedDef{XTestHeader: "default", StatusCode: 202},
			},
			{
				api.HeadersCombinedType4XX,
				&api.HeadersCombined4XX{XTestHeader: "4XX", StatusCode: 404},
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(string(tt.Param), func(t *testing.T) {
				ctx, a, client := create(t)

				r, err := client.HeadersCombined(ctx, api.HeadersCombinedParams{
					Type: tt.Param,
				})
				a.NoError(err, tt.Param)
				a.Equal(tt.Response, r, tt.Param)
			})
		}
	})
}

func TestResponsesOptionalHeaders(t *testing.T) {
	type header struct {
		Name  string
		Value string
	}

	tests := []struct {
		Name    string
		Headers []header
		Error   string
	}{
		{
			Name:    "NoHeaders",
			Headers: nil,
			Error:   `X-Required header: field required`,
		},
		{
			Name: "OnlyOptionalHeaders",
			Headers: []header{
				{"X-Optional", "optional"},
			},
			Error: `X-Required header: field required`,
		},
		{
			Name: "OnlyRequiredHeaders",
			Headers: []header{
				{"X-Required", "required"},
			},
		},
		{
			Name: "AllHeaders",
			Headers: []header{
				{"X-Required", "required"},
				{"X-Optional", "optional"},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			a := require.New(t)

			h := func(w http.ResponseWriter, r *http.Request) {
				for _, h := range tt.Headers {
					w.Header().Set(h.Name, h.Value)
				}
			}
			s := httptest.NewServer(http.HandlerFunc(h))
			t.Cleanup(func() {
				s.Close()
			})

			client, err := api.NewClient(s.URL)
			a.NoError(err)

			_, err = client.OptionalHeaders(context.Background())
			if tt.Error != "" {
				a.ErrorContains(err, tt.Error)
				return
			}
			a.NoError(err)
		})
	}
}

func TestResponsesJSONHeaders(t *testing.T) {
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
				Username: "amongus",
				Role:     api.UserRoleBot,
			},
		},
	}
	raw := jx.Raw(fmt.Sprintf(`{"foo":%q}`, "`\"';,./<>?[]{}\\|~!@#$%^&*()_+-="))
	a, client := testResponsesInit(t, testHTTPResponses{
		headerUser:    user,
		headerJSONRaw: raw,
	})
	ctx := context.Background()

	res, err := client.HeadersJSON(ctx)
	a.NoError(err)

	a.Equal(user, res.XJSONHeader)
	a.Equal(raw, res.XJSONCustomHeader)
}

func TestResponsesPattern(t *testing.T) {
	testData := []byte("bababoi")
	create := func(t *testing.T) (context.Context, *require.Assertions, *api.Client) {
		a, client := testResponsesInit(t, testHTTPResponses{
			data: testData,
		})
		return context.Background(), a, client
	}

	t.Run("IntersectPatternCode", func(t *testing.T) {
		r200 := api.IntersectPatternCodeOKApplicationJSON("200")
		tests := []struct {
			Code int
			Type api.IntersectPatternCodeRes
		}{
			{200, &r200},
			{201, &api.IntersectPatternCode2XXStatusCode{
				StatusCode: 201,
				Response:   201,
			}},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("Code%d", tt.Code), func(t *testing.T) {
				ctx, a, client := create(t)

				r, err := client.IntersectPatternCode(ctx, api.IntersectPatternCodeParams{Code: tt.Code})
				a.NoError(err)
				a.Equal(tt.Type, r)
			})
		}
	})
	t.Run("Combined", func(t *testing.T) {
		tests := []struct {
			Param api.CombinedType
			Type  api.CombinedRes
		}{
			{
				api.CombinedType200,
				&api.CombinedOK{Ok: "200"},
			},
			{
				api.CombinedType2XX,
				&api.Combined2XXStatusCode{
					StatusCode: http.StatusAccepted,
					Response:   http.StatusAccepted,
				},
			},
			{
				api.CombinedType5XX,
				&api.Combined5XXStatusCode{
					StatusCode: http.StatusInternalServerError,
					Response:   true,
				},
			},
			{
				api.CombinedTypeDefault,
				&api.CombinedDefStatusCode{
					StatusCode: http.StatusNotFound,
					Response:   []string{"default"},
				},
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(string(tt.Param), func(t *testing.T) {
				ctx, a, client := create(t)

				r, err := client.Combined(ctx, api.CombinedParams{Type: tt.Param})
				a.NoError(err)
				a.Equal(tt.Type, r)
			})
		}
	})
}

func TestResponseJSONTrailingData(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()

	var response atomic.Value
	response.Store([]byte(`{"ok": "yes"}
{"ok": "trailing"}`))
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(response.Load().([]byte))
	}))
	defer s.Close()

	client, err := api.NewClient(s.URL, api.WithClient(s.Client()))
	a.NoError(err)

	_, err = client.Combined(ctx, api.CombinedParams{Type: api.CombinedType200})
	a.ErrorContains(err, "unexpected trailing data")

	// Trailing lines are ok.
	response.Store([]byte("{\"ok\": \"yes\"}\n\n"))
	resp, err := client.Combined(ctx, api.CombinedParams{Type: api.CombinedType200})
	a.NoError(err)
	a.Equal(&api.CombinedOK{Ok: "yes"}, resp)
}

func TestResponseJSONStreaming(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()

	srv, err := api.NewServer(testHTTPResponses{})
	a.NoError(err)

	s := httptest.NewServer(srv)
	defer s.Close()

	client, err := api.NewClient(s.URL, api.WithClient(s.Client()))
	a.NoError(err)

	r, err := client.StreamJSON(ctx, api.StreamJSONParams{Count: 10})
	a.NoError(err)
	a.IsType(new(api.QueryData), r)
	data := r.(*api.QueryData)
	a.Len(*data, 10)
}

func TestResponseErrorStatusCode(t *testing.T) {
	for _, tt := range []struct {
		code        int
		errContains string
	}{
		{201, "pattern 2XX (code 201)"},
		{501, "pattern 5XX (code 501)"},
		{400, "default (code 400)"},
		{401, "default (code 401)"},
	} {
		tt := tt
		t.Run(fmt.Sprintf("Code%d", tt.code), func(t *testing.T) {
			a := require.New(t)
			ctx := context.Background()

			h := func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(tt.code)
			}
			s := httptest.NewServer(http.HandlerFunc(h))
			t.Cleanup(func() {
				s.Close()
			})

			client, err := api.NewClient(s.URL)
			a.NoError(err)

			_, err = client.Combined(ctx, api.CombinedParams{})
			a.ErrorContains(err, tt.errContains)
			a.ErrorAs(err, new(*validate.InvalidContentTypeError))
		})
	}
}
