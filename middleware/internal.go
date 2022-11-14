package middleware

import "context"

// HookMiddleware is a helper that does ogen request type -> Request type conversion.
//
// NB: this is an internal func, not intended for public use.
func HookMiddleware[RequestType, ParamsType, ResponseType any](
	m Middleware,
	req Request,
	unpack func(Parameters) ParamsType,
	cb func(context.Context, RequestType, ParamsType) (ResponseType, error),
) (r ResponseType, err error) {
	next := func(req Request) (Response, error) {
		var request RequestType
		if body := req.Body; body != nil {
			request = body.(RequestType)
		}
		var params ParamsType
		if unpack != nil {
			params = unpack(req.Params)
		}
		response, err := cb(req.Context, request, params)
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
