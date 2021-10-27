package http

import (
	"net/http"

	"github.com/valyala/fasthttp"
)

type Writer struct {
	Context *fasthttp.RequestCtx
}

func (r Writer) Header() http.Header {
	return http.Header{}
}

func (r Writer) Write(i []byte) (int, error) {
	return r.Context.Write(i)
}

func (r Writer) WriteHeader(statusCode int) {
	r.Context.SetStatusCode(statusCode)
}
