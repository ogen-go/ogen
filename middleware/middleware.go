// Package middleware provides a middleware interface for ogen.
package middleware

import (
	"context"
	"net/http"

	"github.com/ogen-go/ogen/ogenreflect"
)

// ParameterKey is a map key for parameters.
type ParameterKey = ogenreflect.ParameterKey

// Parameters is a map of parameters.
type Parameters = ogenreflect.ParameterMap[any]

// Request is request context type for middleware.
type Request struct {
	// Context is request context.
	Context context.Context
	// OperationName is the ogen operation name. It is guaranteed to be unique and not empty.
	//
	// Deprecated: use Op instead.
	OperationName string
	// OperationID is the spec operation ID, if any.
	//
	// Deprecated: use Op instead.
	OperationID string
	// Op contains the operation information.
	Op ogenreflect.RuntimeOperation
	// Body is the operation request body. May be nil, if the operation has not body.
	Body any
	// Params is the operation parameters.
	Params Parameters
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
