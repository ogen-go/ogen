package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (r roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return r(req)
}

func Test_parseSpecPath(t *testing.T) {
	testdata := []byte(`{}`)
	urlPath := func(p string) *url.URL {
		return &url.URL{Path: p}
	}
	urlParse := func(s string) *url.URL {
		u, err := url.Parse(s)
		require.NoError(t, err)
		return u
	}

	type testCase struct {
		input        string
		httpData     []byte
		fileData     []byte
		wantFilename string
		wantURL      *url.URL
	}

	tests := []testCase{
		{"spec.json", nil, testdata, "spec.json", urlPath("spec.json")},
		{"./spec.json", nil, testdata, "spec.json", urlPath("spec.json")},
		{"_testdata/spec.json", nil, testdata, "spec.json", urlPath("_testdata/spec.json")},

		{"http://example.com/spec.json", testdata, nil, "spec.json", urlParse("http://example.com/spec.json")},
	}
	if runtime.GOOS == "windows" {
		tests = append(tests, []testCase{
			{`_testdata\spec.json`, nil, testdata, "spec.json", urlPath("_testdata/spec.json")},
			{`C:\_testdata\spec.json`, nil, testdata, "spec.json", urlPath("C:/_testdata/spec.json")},
		}...)
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			var (
				data     = tt.httpData
				readFile = func(filename string) ([]byte, error) {
					return nil, errors.Errorf("unexpected read file: %q", filename)
				}
				httpClient = &http.Client{
					Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewReader(data)),
						}, nil
					}),
				}
			)
			if data == nil {
				data = tt.fileData
				readFile = func(filename string) ([]byte, error) {
					return data, nil
				}
				httpClient = &http.Client{
					Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
						return nil, errors.Errorf("unexpected http request: %q", req.URL)
					}),
				}
			}

			f, _, err := parseSpecPath(tt.input, httpClient, readFile, zaptest.NewLogger(t))
			a.NoError(err)
			a.Equal(tt.wantFilename, f.fileName)
			a.Equal(tt.wantURL, f.rootURL)
		})
	}
}
