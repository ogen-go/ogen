// Package http implements crazy ideas for http optimizations that should be
// mostly std compatible.
package http

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// NewRequest creates a new http.Request.
func NewRequest(ctx context.Context, method string, u *url.URL, body io.Reader) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, method, u.String(), body)
}

// SetBody sets request body.
func SetBody(req *http.Request, body io.Reader, contentType string) {
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Body = io.NopCloser(body)

	switch v := body.(type) {
	case *bytes.Buffer:
		req.ContentLength = int64(v.Len())
		buf := v.Bytes()
		req.GetBody = func() (io.ReadCloser, error) {
			r := bytes.NewReader(buf)
			return io.NopCloser(r), nil
		}
	case *bytes.Reader:
		req.ContentLength = int64(v.Len())
		snapshot := *v
		req.GetBody = func() (io.ReadCloser, error) {
			r := snapshot
			return io.NopCloser(&r), nil
		}
	case *strings.Reader:
		req.ContentLength = int64(v.Len())
		snapshot := *v
		req.GetBody = func() (io.ReadCloser, error) {
			r := snapshot
			return io.NopCloser(&r), nil
		}
	default:
	}
}
