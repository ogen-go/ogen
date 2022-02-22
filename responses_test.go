package ogen_test

import (
	"bytes"
	"context"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/internal/test_http_responses"
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

func TestResponses(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()

	testData := []byte("bababoi")
	srv, err := api.NewServer(testHTTPResponses{
		data: testData,
	})
	a.NoError(err)

	s := httptest.NewServer(srv)
	defer s.Close()

	client, err := api.NewClient(s.URL, api.WithClient(s.Client()))
	a.NoError(err)

	{
		r, err := client.AnyContentTypeBinaryStringSchema(ctx)
		a.NoError(err)
		data, err := io.ReadAll(r.Data)
		a.NoError(err)
		a.Equal(testData, data)
	}
	{
		r, err := client.AnyContentTypeBinaryStringSchemaDefault(ctx)
		a.NoError(err)
		data, err := io.ReadAll(r.Response.Data)
		a.NoError(err)
		a.Equal(testData, data)
		a.Equal(200, r.StatusCode)
	}
	{
		r, err := client.MultipleGenericResponses(ctx)
		a.NoError(err)
		a.Equal(&api.NilString{Null: true}, r)
	}
	{
		r, err := client.OctetStreamBinaryStringSchema(ctx)
		a.NoError(err)
		data, err := io.ReadAll(r.Data)
		a.NoError(err)
		a.Equal(testData, data)
	}
	{
		r, err := client.OctetStreamEmptySchema(ctx)
		a.NoError(err)
		data, err := io.ReadAll(r.Data)
		a.NoError(err)
		a.Equal(testData, data)
	}
}
