package ogen

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ht "github.com/ogen-go/ogen/http"
	api "github.com/ogen-go/ogen/internal/sample_api"
	"github.com/ogen-go/ogen/validate"
)

func testForm() api.TestForm {
	return api.TestForm{
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
	}
}

func checkTestFormValues(a *assert.Assertions, form url.Values) {
	a.Equal("10", form.Get("id"))
	a.Equal("00000000-0000-0000-0000-000000000000", form.Get("uuid"))
	a.Equal("foobar", form.Get("description"))
	a.Equal([]string{"foo", "bar"}, form["array"])
	a.Equal("10", form.Get("min"))
	a.Equal("10", form.Get("max"))
	a.Equal("10", form.Get("deepObject[min]"))
	a.Equal("10", form.Get("deepObject[max]"))
}

type testFormServer struct {
	a *assert.Assertions
	*sampleAPIServer
}

func (s testFormServer) TestFormURLEncoded(ctx context.Context, req api.TestForm) (r api.TestFormURLEncodedOK, _ error) {
	s.a.Equal(testForm(), req)
	return r, nil
}

func (s testFormServer) TestMultipart(ctx context.Context, req api.TestForm) (r api.TestMultipartOK, _ error) {
	s.a.Equal(testForm(), req)
	return r, nil
}

func (s testFormServer) TestMultipartUpload(ctx context.Context, req api.TestMultipartUploadReq) (r string, _ error) {
	f := req.FileName
	if val := f.Header.Get("Content-Type"); val != "image/jpeg" {
		return "", validate.InvalidContentType(val)
	}

	var b strings.Builder
	_, err := io.Copy(&b, f.File)
	return b.String(), err
}

func TestURIEncodingE2E(t *testing.T) {
	a := assert.New(t)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	handler := &testFormServer{
		a:               a,
		sampleAPIServer: new(sampleAPIServer),
	}
	apiServer, err := api.NewServer(handler, handler)
	require.NoError(t, err)

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		a.NoError(req.ParseForm())
		checkTestFormValues(a, req.PostForm)
		apiServer.ServeHTTP(w, req)
	}))
	defer s.Close()

	client, err := api.NewClient(s.URL, handler)
	require.NoError(t, err)

	_, err = client.TestFormURLEncoded(ctx, testForm())
	a.NoError(err)
}

func TestMultipartEncodingE2E(t *testing.T) {
	a := assert.New(t)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	handler := &testFormServer{
		a:               a,
		sampleAPIServer: new(sampleAPIServer),
	}
	apiServer, err := api.NewServer(handler, handler)
	require.NoError(t, err)

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		a.NoError(req.ParseMultipartForm(32 << 20))
		form := url.Values(req.MultipartForm.Value)
		checkTestFormValues(a, form)
		apiServer.ServeHTTP(w, req)
	}))
	defer s.Close()

	client, err := api.NewClient(s.URL, handler)
	require.NoError(t, err)

	_, err = client.TestMultipart(ctx, testForm())
	a.NoError(err)
}

func TestMultipartUploadE2E(t *testing.T) {
	a := assert.New(t)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	handler := &testFormServer{
		a:               a,
		sampleAPIServer: new(sampleAPIServer),
	}
	apiServer, err := api.NewServer(handler, handler)
	require.NoError(t, err)

	s := httptest.NewServer(apiServer)
	defer s.Close()

	client, err := api.NewClient(s.URL, handler)
	require.NoError(t, err)

	data := "data"
	r, err := client.TestMultipartUpload(ctx, api.TestMultipartUploadReq{
		FileName: ht.MultipartFile{
			Name: "pablo.jpg",
			File: strings.NewReader(data),
			Header: textproto.MIMEHeader{
				"Content-Type": []string{"image/jpeg"},
			},
		},
	})
	a.NoError(err)
	a.Equal(data, r)
}
