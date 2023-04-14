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

	"golang.org/x/sync/errgroup"
)

// Client represents http client.
type Client interface {
	Do(r *http.Request) (*http.Response, error)
}

// NewRequest creates a new http.Request.
func NewRequest(ctx context.Context, method string, u *url.URL) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, method, u.String(), http.NoBody)
}

func initRequest(req *http.Request, contentType string) {
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
}

// SetBody sets request body.
func SetBody(req *http.Request, body io.Reader, contentType string) {
	initRequest(req, contentType)
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
	}
}

// SetCloserBody sets request body which should be closed after request.
func SetCloserBody(req *http.Request, body io.ReadCloser, contentType string) {
	initRequest(req, contentType)
	req.Body = body
}

// CreateBodyWriter is a helper to create a reader from a writer body.
func CreateBodyWriter(cb func(w io.Writer) error) io.ReadCloser {
	piper, pipew := io.Pipe()

	wg := new(errgroup.Group)
	wg.Go(func() (rerr error) {
		defer func() {
			if rerr != nil {
				_ = pipew.CloseWithError(rerr)
			} else {
				_ = pipew.Close()
			}
		}()
		return cb(pipew)
	})

	return bodyReader{
		r:  piper,
		w:  pipew,
		wg: wg,
	}
}

type bodyReader struct {
	r  *io.PipeReader
	w  *io.PipeWriter
	wg *errgroup.Group
}

func (w bodyReader) Read(p []byte) (int, error) {
	return w.r.Read(p)
}

func (w bodyReader) Close() (rerr error) {
	rerr = w.r.Close()
	_ = w.wg.Wait()
	return rerr
}
