package ogen

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/sample_api"
)

func TestURIEncodingE2E(t *testing.T) {
	a := assert.New(t)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		a.NoError(req.ParseForm())
		a.Equal("10", req.PostForm.Get("id"))
		a.Equal("00000000-0000-0000-0000-000000000000", req.PostForm.Get("uuid"))
		a.Equal("foobar", req.PostForm.Get("description"))
		a.Equal([]string{"foo", "bar"}, req.PostForm["array"])
		a.Equal("10", req.PostForm.Get("min"))
		a.Equal("10", req.PostForm.Get("max"))
		a.Equal("10", req.PostForm.Get("deepObject[min]"))
		a.Equal("10", req.PostForm.Get("deepObject[max]"))
		w.WriteHeader(http.StatusOK)
	}))
	defer s.Close()

	client, err := api.NewClient(s.URL, &sampleAPIServer{})
	require.NoError(t, err)

	_, err = client.TestFormURLEncoded(ctx, api.TestForm{
		ID:          api.NewOptInt(10),
		UUID:        api.NewOptUUID(uuid.MustParse("00000000-0000-0000-0000-000000000000")),
		Description: "foobar",
		Array:       []string{"foo", "bar"},
		Object: api.NewOptTestFormObject(api.TestFormObject{
			Min: api.NewOptInt(10),
			Max: 10,
		}),
		DeepObject: api.NewOptTestFormDeepObject(api.TestFormDeepObject{
			Min: api.NewOptInt(10),
			Max: 10,
		}),
	})
	a.NoError(err)
}
