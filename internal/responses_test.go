package internal_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/test_http_responses"
)

type testHTTPResponses struct {
	data []byte
}

func (t testHTTPResponses) AnyContentTypeBinaryStringSchema(ctx context.Context) (api.AnyContentTypeBinaryStringSchemaOK, error) {
	return api.AnyContentTypeBinaryStringSchemaOK{
		Data: bytes.NewReader(t.data),
	}, nil
}

func (t testHTTPResponses) AnyContentTypeBinaryStringSchemaDefault(ctx context.Context) (api.AnyContentTypeBinaryStringSchemaDefaultDefStatusCode, error) {
	return api.AnyContentTypeBinaryStringSchemaDefaultDefStatusCode{
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

func (t testHTTPResponses) Headers200(ctx context.Context) (api.Headers200OK, error) {
	return api.Headers200OK{
		TestHeader: "foo",
	}, nil
}

func (t testHTTPResponses) HeadersDefault(ctx context.Context) (api.HeadersDefaultDef, error) {
	return api.HeadersDefaultDef{
		TestHeader: "202",
		StatusCode: 202,
	}, nil
}

func (t testHTTPResponses) HeadersPattern(ctx context.Context) (api.HeadersPattern4XX, error) {
	return api.HeadersPattern4XX{
		TestHeader: "404",
		StatusCode: 404,
	}, nil
}

func (t testHTTPResponses) HeadersCombined(ctx context.Context, params api.HeadersCombinedParams) (api.HeadersCombinedRes, error) {
	switch params.Type {
	case api.HeadersCombinedType200:
		return &api.HeadersCombinedOK{
			TestHeader: "200",
		}, nil
	case api.HeadersCombinedTypeDefault:
		return &api.HeadersCombinedDef{
			TestHeader: "default",
			StatusCode: 202,
		}, nil
	case api.HeadersCombinedType4XX:
		return &api.HeadersCombined4XX{
			TestHeader: "4XX",
			StatusCode: 404,
		}, nil
	default:
		panic(fmt.Sprintf("unknown type %q", params.Type))
	}
}

func testResponsesInit(t *testing.T, testData []byte) (*require.Assertions, *api.Client) {
	a := require.New(t)

	srv, err := api.NewServer(testHTTPResponses{
		data: testData,
	})
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
		a, client := testResponsesInit(t, testData)
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
		a, client := testResponsesInit(t, testData)
		return context.Background(), a, client
	}

	t.Run("Headers200", func(t *testing.T) {
		ctx, a, client := create(t)

		r, err := client.Headers200(ctx)
		a.NoError(err)
		a.Equal(r.TestHeader, "foo")
	})
	t.Run("HeadersDefault", func(t *testing.T) {
		ctx, a, client := create(t)

		r, err := client.HeadersDefault(ctx)
		a.NoError(err)
		a.Equal(r.StatusCode, 202)
		a.Equal(r.TestHeader, "202")
	})
	t.Run("HeadersPattern", func(t *testing.T) {
		ctx, a, client := create(t)

		r, err := client.HeadersPattern(ctx)
		a.NoError(err)
		a.Equal(r.StatusCode, 404)
		a.Equal(r.TestHeader, "404")
	})
	t.Run("HeadersCombined", func(t *testing.T) {
		tests := []struct {
			Param    api.HeadersCombinedType
			Response api.HeadersCombinedRes
		}{
			{
				api.HeadersCombinedType200,
				&api.HeadersCombinedOK{TestHeader: "200"},
			},
			{
				api.HeadersCombinedTypeDefault,
				&api.HeadersCombinedDef{TestHeader: "default", StatusCode: 202},
			},
			{
				api.HeadersCombinedType4XX,
				&api.HeadersCombined4XX{TestHeader: "4XX", StatusCode: 404},
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

func TestResponsesPattern(t *testing.T) {
	testData := []byte("bababoi")
	create := func(t *testing.T) (context.Context, *require.Assertions, *api.Client) {
		a, client := testResponsesInit(t, testData)
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
