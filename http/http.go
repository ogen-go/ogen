// Package http implements crazy ideas for http optimizations that should be
// mostly std compatible.
package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"unsafe"

	"github.com/go-faster/errors"
)

// httpRequest is copied version of http.Request structure.
type httpRequest struct {
	Method           string
	Proto            string // "HTTP/1.0"
	URL              *url.URL
	ProtoMajor       int // 1
	ProtoMinor       int // 0
	Header           http.Header
	Body             io.ReadCloser
	GetBody          func() (io.ReadCloser, error)
	ContentLength    int64
	TransferEncoding []string
	Close            bool
	Host             string
	Form             url.Values
	PostForm         url.Values
	MultipartForm    *multipart.Form
	Trailer          http.Header
	RemoteAddr       string
	RequestURI       string
	TLS              *tls.ConnectionState
	Cancel           <-chan struct{}
	Response         *http.Response
	ctx              context.Context
}

func toPrivate(req *http.Request) *httpRequest {
	return (*httpRequest)(unsafe.Pointer(req))
}

// Set sets request context without shallow copy of request.
func Set(req *http.Request, ctx context.Context) {
	p := toPrivate(req)
	p.ctx = ctx
}

// SetValue wraps context.WithValue call on request context.
func SetValue(req *http.Request, k, v interface{}) {
	ctx := context.WithValue(req.Context(), k, v)
	Set(req, ctx)
}

func init() {
	// Explicitly check that structures have at least equal size.
	stdSize := unsafe.Sizeof(http.Request{})
	gotSize := unsafe.Sizeof(httpRequest{})
	if stdSize != gotSize {
		panic(errors.Errorf("%d (net/http) != %d", stdSize, gotSize))
	}
}

// Given a string of the form "host", "host:port", or "[ipv6::address]:port",
// return true if the string includes a port.
func hasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }

// removeEmptyPort strips the empty port in ":port" to ""
// as mandated by RFC 3986 Section 6.2.3.
func removeEmptyPort(host string) string {
	if hasPort(host) {
		return strings.TrimSuffix(host, ":")
	}
	return host
}

// NewRequest is optimized version of http.NewRequestWithContext.
func NewRequest(ctx context.Context, method string, u *url.URL, body io.Reader) *http.Request {
	req := new(http.Request)
	Set(req, ctx)

	rc, ok := body.(io.ReadCloser)
	if !ok && body != nil {
		rc = io.NopCloser(body)
	}

	// The host's colon:port should be normalized. See Issue 14836.
	u.Host = removeEmptyPort(u.Host)

	req.Proto = "HTTP/1.1"
	req.ProtoMajor = 1
	req.ProtoMinor = 1
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	req.Body = rc
	req.Host = u.Host
	req.Method = method
	req.URL = u

	if body != nil {
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
		if req.GetBody != nil && req.ContentLength == 0 {
			req.Body = http.NoBody
			req.GetBody = func() (io.ReadCloser, error) { return http.NoBody, nil }
		}
	}

	return req
}

// SetBody sets request body.
func SetBody(req *http.Request, body io.Reader, contentType string) {
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	req.Header.Set("Content-Type", contentType)
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
