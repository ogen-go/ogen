// Package middleware provides a middleware interface for ogen.
package middleware

import (
	"context"
	"net/http"

	"github.com/ogen-go/ogen/openapi"
)

// ParameterKey is a map key for parameters.
type ParameterKey struct {
	// Name is the name of the parameter.
	Name string
	// In is the location of the parameter.
	In openapi.ParameterLocation
}

// Parameters is a map of parameters.
type Parameters map[ParameterKey]any

func (p Parameters) find(name string, in openapi.ParameterLocation) (v any, ok bool) {
	v, ok = p[ParameterKey{Name: name, In: in}]
	return v, ok
}

// Query returns a parameter from the query.
func (p Parameters) Query(name string) (any, bool) {
	return p.find(name, openapi.LocationQuery)
}

// Header returns a parameter from the header.
func (p Parameters) Header(name string) (any, bool) {
	return p.find(name, openapi.LocationHeader)
}

// Path returns a parameter from the path.
func (p Parameters) Path(name string) (any, bool) {
	return p.find(name, openapi.LocationPath)
}

// Cookie returns a parameter from the cookie.
func (p Parameters) Cookie(name string) (any, bool) {
	return p.find(name, openapi.LocationCookie)
}

// Request is request context type for middleware.
type Request struct {
	// Context is request context.
	Context context.Context
	// OperationName is the ogen operation name. It is guaranteed to be unique and not empty.
	OperationName string
	// OperationSummary is the ogen operation summary.
	OperationSummary string
	// OperationID is the spec operation ID, if any.
	OperationID string
	// Body is the operation request body. May be nil, if the operation has not body.
	Body any
	// Params is the operation parameters.
	Params Parameters
	// Raw is the raw http request.
	Raw *http.Request
}

// SetContext sets Context in Request.
func (r *Request) SetContext(ctx context.Context) {
	r.Context = ctx
}

// Response is response type for middleware.
type Response struct {
	// Type is the operation response type.
	Type any
}

type (
	// Next is the next middleware/handler in the chain.
	Next = func(req Request) (Response, error)
	// Middleware is middleware type.
	Middleware func(req Request, next Next) (Response, error)
)

// ChainMiddlewares chains middlewares into a single middleware, which will be executed in the order they are passed.
func ChainMiddlewares(m ...Middleware) Middleware {
	if len(m) == 0 {
		return func(req Request, next Next) (Response, error) {
			return next(req)
		}
	}
	tail := ChainMiddlewares(m[1:]...)
	return func(req Request, next Next) (Response, error) {
		return m[0](req, func(req Request) (Response, error) {
			return tail(req, next)
		})
	}
}
