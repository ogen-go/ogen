// Package middleware provides a middleware interface for ogen.
package middleware

import (
	"context"
	"net/http"
)

// Request is request context type for middleware.
type Request struct {
	// Context is request context.
	Context context.Context
	// OperationName is the ogen operation name. It is guaranteed to be unique and not empty.
	OperationName string
	// OperationID is the spec operation ID, if any.
	OperationID string
	// Body is the operation request body. May be nil, if the operation has not body.
	Body any
	// Params is the operation parameters.
	Params map[string]any
	// Raw is the raw http request.
	Raw *http.Request
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
