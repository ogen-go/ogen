package http

import (
	"net/http"
	"sync"
)

var reqPool = sync.Pool{
	New: func() interface{} {
		return new(http.Request)
	},
}

// AcquireRequest returns new *http.Request from pool.
func AcquireRequest() *http.Request {
	return reqPool.Get().(*http.Request)
}

// PutRequest resets *http.Request and puts to pool.
func PutRequest(r *http.Request) {
	// Reset public.
	r.Body = nil
	r.ContentLength = 0
	r.Form = nil
	r.MultipartForm = nil
	r.PostForm = nil

	// Reset URL with pool.
	if u := r.URL; u != nil {
		r.URL = nil
		PutURL(u)
	}

	r.RequestURI = ""
	r.RemoteAddr = ""
	r.Response = nil
	r.TLS = nil
	r.Trailer = nil
	r.TransferEncoding = nil
	r.Proto = ""
	r.ProtoMajor = 0
	r.ProtoMinor = 0
	r.Cancel = nil

	// Reusing header map.
	if len(r.Header) > 0 {
		for k := range r.Header {
			delete(r.Header, k)
		}
	}

	// Reset internal.
	Set(r, nil)

	reqPool.Put(r)
}
