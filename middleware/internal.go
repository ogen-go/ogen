package middleware

import "context"

// HookMiddleware is a helper that does ogen request type -> Request type conversion.
//
// NB: this is an internal func, not intended for public use.
func HookMiddleware[RequestType, ParamsType, ResponseType any](
	m Middleware,
	params ParamsType,
	req Request,
	cb func(context.Context, ParamsType, RequestType) (ResponseType, error),
) (r ResponseType, err error) {
	next := func(req Request) (Response, error) {
		request := req.Body.(RequestType)
		response, err := cb(req.Context, params, request)
		if err != nil {
			return Response{}, err
		}
		return Response{Type: response}, nil
	}
	resp, err := m(req, next)
	if err != nil {
		return r, err
	}
	return resp.Type.(ResponseType), nil
}
