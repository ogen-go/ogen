package ogen

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/sample_api"
	"github.com/ogen-go/ogen/uri"
)

func decodeURI(input url.Values, r *api.URIStruct) error {
	q := uri.NewQueryDecoder(input)
	cfg := uri.QueryParameterDecodingConfig{
		Name:    "",
		Style:   uri.QueryStyleForm,
		Explode: true,
		Fields: []uri.QueryParameterObjectField{
			{"id", false},
			{"uuid", false},
			{"description", true},
		},
	}
	if err := q.HasParam(cfg); err != nil {
		return err
	}
	return q.DecodeParam(cfg, func(d uri.Decoder) error {
		return r.DecodeURI(d)
	})
}

func TestURIStruct(t *testing.T) {
	tests := []struct {
		Input    url.Values
		Expected api.URIStruct
		Error    bool
	}{
		{
			url.Values{
				"description": {"foobar"},
			},
			api.URIStruct{
				Description: "foobar",
			},
			false,
		},
		{
			url.Values{
				"id":          {"10"},
				"description": {"foobar"},
			},
			api.URIStruct{
				ID:          api.NewOptInt(10),
				Description: "foobar",
			},
			false,
		},
		{
			url.Values{
				"id":          {"10"},
				"uuid":        {"00000000-0000-0000-0000-000000000000"},
				"description": {"foobar"},
			},
			api.URIStruct{
				ID:          api.NewOptInt(10),
				UUID:        api.NewOptUUID(uuid.MustParse("00000000-0000-0000-0000-000000000000")),
				Description: "foobar",
			},
			false,
		},
		{
			url.Values{
				"id":          {"foobar"},
				"description": {"foobar"},
			},
			api.URIStruct{},
			true,
		},
		{
			url.Values{},
			api.URIStruct{},
			true,
		},
	}
	for i, tc := range tests {
		// Make range value copy to prevent data races.
		tc := tc
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			r := api.URIStruct{}
			if err := decodeURI(tc.Input, &r); tc.Error {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.Expected, r)
			}
		})
	}
}

func TestE2EClient(t *testing.T) {
	a := assert.New(t)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		a.NoError(req.ParseForm())
		a.Equal("10", req.PostForm.Get("id"))
		a.Equal("00000000-0000-0000-0000-000000000000", req.PostForm.Get("uuid"))
		a.Equal("foobar", req.PostForm.Get("description"))
		w.WriteHeader(http.StatusOK)
	}))
	defer s.Close()

	client, err := api.NewClient(s.URL, &sampleAPIServer{})
	require.NoError(t, err)

	_, err = client.TestFormURLEncoded(ctx, api.URIStruct{
		ID:          api.NewOptInt(10),
		UUID:        api.NewOptUUID(uuid.MustParse("00000000-0000-0000-0000-000000000000")),
		Description: "foobar",
	})
	a.NoError(err)
}
