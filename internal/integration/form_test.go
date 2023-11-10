package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ht "github.com/ogen-go/ogen/http"
	api "github.com/ogen-go/ogen/internal/integration/test_form"
	"github.com/ogen-go/ogen/validate"
)

func testForm() *api.TestForm {
	return &api.TestForm{
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

func testFormMultipart() *api.TestFormMultipart {
	return &api.TestFormMultipart{
		ID:          api.NewOptInt(10),
		UUID:        api.NewOptUUID(uuid.MustParse("00000000-0000-0000-0000-000000000000")),
		Description: "foobar",
		Array:       []string{"foo", "bar"},
		Object: api.NewOptTestFormMultipartObject(api.TestFormMultipartObject{
			Min: api.NewOptInt(10),
			Max: 10,
		}),
		DeepObject: api.NewOptTestFormMultipartDeepObject(api.TestFormMultipartDeepObject{
			Min: api.NewOptInt(10),
			Max: 10,
		}),
	}
}

func checkTestFormValues(a *assert.Assertions, form url.Values, multipartForm bool) {
	a.Equal("10", form.Get("id"))
	a.Equal("00000000-0000-0000-0000-000000000000", form.Get("uuid"))
	a.Equal("foobar", form.Get("description"))
	a.Equal([]string{"foo", "bar"}, form["array"])
	if multipartForm {
		a.JSONEq(`{"min":10,"max":10}`, form.Get("object"))
	} else {
		a.Equal("10", form.Get("min"))
		a.Equal("10", form.Get("max"))
	}
	a.Equal("10", form.Get("deepObject[min]"))
	a.Equal("10", form.Get("deepObject[max]"))
}

var _ api.Handler = (*testFormServer)(nil)

type testFormServer struct {
	a *assert.Assertions
}

func (s testFormServer) TestFormURLEncoded(ctx context.Context, req *api.TestForm) error {
	s.a.Equal(testForm(), req)
	return nil
}

func (s testFormServer) TestMultipart(ctx context.Context, req *api.TestFormMultipart) error {
	s.a.Equal(testFormMultipart(), req)
	return nil
}

func (s testFormServer) TestMultipartUpload(ctx context.Context, req *api.TestMultipartUploadReq) (
	*api.TestMultipartUploadOK,
	error,
) {
	r := new(api.TestMultipartUploadOK)
	readFile := func(f ht.MultipartFile, to *string) error {
		var b strings.Builder
		if _, err := io.Copy(&b, f.File); err != nil {
			return err
		}
		*to = b.String()
		return nil
	}

	f := req.File
	if val := f.Header.Get("Content-Type"); val != "image/jpeg" {
		return r, validate.InvalidContentType(val)
	}

	if err := readFile(req.File, &r.File); err != nil {
		return r, errors.Wrap(err, "file")
	}
	if file, ok := req.OptionalFile.Get(); ok {
		r.OptionalFile.Set = true
		if err := readFile(file, &r.OptionalFile.Value); err != nil {
			return r, errors.Wrap(err, "optional_file")
		}
	}
	if err := func() error {
		for idx, file := range req.Files {
			var val string
			if err := readFile(file, &val); err != nil {
				return errors.Wrapf(err, "[%d]", idx)
			}
			r.Files = append(r.Files, val)
		}
		return nil
	}(); err != nil {
		return r, errors.Wrap(err, "files")
	}

	return r, nil
}

func (s testFormServer) TestReuseFormOptionalSchema(ctx context.Context, req api.OptSharedRequestMultipart) error {
	return nil
}

func (s testFormServer) TestReuseFormSchema(ctx context.Context, req *api.SharedRequestMultipart) error {
	return nil
}

func (s testFormServer) TestShareFormSchema(ctx context.Context, req api.TestShareFormSchemaReq) error {
	return nil
}

func (s testFormServer) OnlyForm(ctx context.Context, req *api.OnlyFormReq) error {
	return nil
}

func (s testFormServer) OnlyMultipartFile(ctx context.Context, req *api.OnlyMultipartFileReq) error {
	return nil
}

func (s testFormServer) OnlyMultipartForm(ctx context.Context, req *api.OnlyMultipartFormReq) error {
	return nil
}

func TestURIEncodingE2E(t *testing.T) {
	tests := []struct {
		name        string
		serverSetup func(h http.Handler) *httptest.Server
	}{
		{
			`Plain`,
			httptest.NewServer,
		},
		{
			`Redirect`,
			func(h http.Handler) *httptest.Server {
				mux := http.NewServeMux()
				mux.HandleFunc("/testFormURLEncoded", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Location", "/redirectTo")
					w.WriteHeader(http.StatusPermanentRedirect)
				})
				mux.HandleFunc("/redirectTo", func(w http.ResponseWriter, r *http.Request) {
					// Overwrite the request URI for ogen handler.
					r.URL = &url.URL{Path: "/testFormURLEncoded"}
					h.ServeHTTP(w, r)
				})
				return httptest.NewServer(mux)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
			defer cancel()

			handler := &testFormServer{
				a: a,
			}
			apiServer, err := api.NewServer(handler)
			require.NoError(t, err)

			s := tt.serverSetup(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				a.NoError(req.ParseForm())
				checkTestFormValues(a, req.PostForm, false)
				apiServer.ServeHTTP(w, req)
			}))
			defer s.Close()

			client, err := api.NewClient(s.URL)
			require.NoError(t, err)

			err = client.TestFormURLEncoded(ctx, testForm())
			a.NoError(err)
		})
	}
}

func TestMultipartEncodingE2E(t *testing.T) {
	a := assert.New(t)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	handler := &testFormServer{
		a: a,
	}
	apiServer, err := api.NewServer(handler)
	require.NoError(t, err)

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		a.NoError(req.ParseMultipartForm(32 << 20))
		form := url.Values(req.MultipartForm.Value)
		checkTestFormValues(a, form, true)
		apiServer.ServeHTTP(w, req)
	}))
	defer s.Close()

	client, err := api.NewClient(s.URL)
	require.NoError(t, err)

	err = client.TestMultipart(ctx, testFormMultipart())
	a.NoError(err)
}

func TestMultipartUploadE2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	handler := &testFormServer{
		a: assert.New(t),
	}
	// Set low limit for BigFile test.
	apiServer, err := api.NewServer(handler,
		api.WithMaxMultipartMemory(1024),
	)
	require.NoError(t, err)

	s := httptest.NewServer(apiServer)
	defer s.Close()

	client, err := api.NewClient(s.URL)
	require.NoError(t, err)

	tests := []struct {
		name     string
		file     string
		optional api.OptString
		array    []string
		wantErr  bool
	}{
		{name: "OnlyFile", file: "data"},
		// Ensure correct handling of file which is bigger than max memory limit.
		{name: "BigFile", file: strings.Repeat("a", 16384)},
		{name: "All", file: "file", optional: api.NewOptString("optional"), array: []string{"1", "2"}},

		{name: "TooBigArray", file: "file", array: []string{"1", "2", "3", "4", "5", "6"}, wantErr: true},
	}
	for _, tt := range tests {
		if tt.array == nil {
			tt.array = []string{}
		}
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			a := require.New(t)

			req := &api.TestMultipartUploadReq{
				File: ht.MultipartFile{
					Name: "pablo.jpg",
					File: strings.NewReader(tt.file),
					Header: textproto.MIMEHeader{
						"Content-Type": []string{"image/jpeg"},
					},
				},
			}
			if val, ok := tt.optional.Get(); ok {
				req.OptionalFile = api.NewOptMultipartFile(ht.MultipartFile{
					Name: "henry.jpg",
					File: strings.NewReader(val),
				})
			}
			for idx, val := range tt.array {
				req.Files = append(req.Files, ht.MultipartFile{
					Name: fmt.Sprintf("file%d.png", idx),
					File: strings.NewReader(val),
				})
			}

			r, err := client.TestMultipartUpload(ctx, req)
			if tt.wantErr {
				a.Error(err)
				return
			}
			a.NoError(err)
			a.Equal(tt.file, r.File)
			a.Equal(tt.optional, r.OptionalFile)
			a.Equal(tt.array, r.Files)

			// FIXME(tdakkota): Check metrics. OpenTelemetry removed metrictest package.
		})
	}
}

func TestFormAdditionalProps(t *testing.T) {
	handler := &testFormServer{
		a: assert.New(t),
	}
	apiServer, err := api.NewServer(handler)
	require.NoError(t, err)

	s := httptest.NewServer(apiServer)
	defer s.Close()

	client := s.Client()
	t.Run("OnlyForm", func(t *testing.T) {
		resp, err := client.PostForm(s.URL+"/onlyForm", url.Values{
			"definitely-not-defined": []string{"value"},
			"field":                  []string{"10"},
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		var jsonErr struct {
			Message string `json:"error_message"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&jsonErr))
		require.Contains(t, jsonErr.Message, `unexpected field "definitely-not-defined"`)
	})

	sendMultipart := func(url string, cb func(*multipart.Writer)) (*http.Response, error) {
		req, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			return nil, err
		}

		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		cb(mw)
		require.NoError(t, mw.Close())

		req.Header.Add("Content-Type", mw.FormDataContentType())
		req.Body = io.NopCloser(&buf)

		return client.Do(req)
	}
	t.Run("OnlyMultipartForm", func(t *testing.T) {
		resp, err := sendMultipart(s.URL+"/onlyMultipartForm", func(mw *multipart.Writer) {
			mw.WriteField("definitely-not-defined", "value")
			mw.WriteField("field", "10")
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		var jsonErr struct {
			Message string `json:"error_message"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&jsonErr))
		require.Contains(t, jsonErr.Message, `unexpected field "definitely-not-defined"`)
	})

	t.Run("OnlyMultipartFile", func(t *testing.T) {
		resp, err := sendMultipart(s.URL+"/onlyMultipartFile", func(mw *multipart.Writer) {
			mw.WriteField("definitely-not-defined", "value")
			f, err := mw.CreateFormFile("file", "pablo.txt")
			require.NoError(t, err)
			io.WriteString(f, "text")
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		var jsonErr struct {
			Message string `json:"error_message"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&jsonErr))
		require.Contains(t, jsonErr.Message, `unexpected field "definitely-not-defined"`)
	})
}
