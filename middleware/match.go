package middleware

import (
	"regexp"

	"github.com/ogen-go/ogen/internal/xmaps"
)

// OperationID calls the next middleware if request operation ID matches the given operationID.
func OperationID(m Middleware, operationID ...string) Middleware {
	switch len(operationID) {
	case 0:
		return justCallNext
	case 1:
		val := operationID[0]
		return func(req Request, next Next) (Response, error) {
			if req.OperationID == val {
				return m(req, next)
			}
			return next(req)
		}
	default:
		set := xmaps.BuildSet(operationID...)
		return func(req Request, next Next) (Response, error) {
			if _, ok := set[req.OperationID]; ok {
				return m(req, next)
			}
			return next(req)
		}
	}
}

// OperationName calls the next middleware if request operation name matches the given operationName.
func OperationName(m Middleware, operationName ...string) Middleware {
	switch len(operationName) {
	case 0:
		return justCallNext
	case 1:
		val := operationName[0]
		return func(req Request, next Next) (Response, error) {
			if req.OperationName == val {
				return m(req, next)
			}
			return next(req)
		}
	default:
		set := xmaps.BuildSet(operationName...)
		return func(req Request, next Next) (Response, error) {
			if _, ok := set[req.OperationName]; ok {
				return m(req, next)
			}
			return next(req)
		}
	}
}

// PathRegex calls the next middleware if request path matches the given regex.
func PathRegex(re *regexp.Regexp, m Middleware) Middleware {
	if re == nil {
		return justCallNext
	}

	return func(req Request, next Next) (Response, error) {
		if re.MatchString(req.Raw.URL.Path) {
			return m(req, next)
		}
		return next(req)
	}
}

// BodyType calls the next middleware if request body type matches the given type.
func BodyType[T any](m Middleware) Middleware {
	return func(req Request, next Next) (Response, error) {
		if _, ok := req.Body.(T); ok {
			return m(req, next)
		}
		return next(req)
	}
}
