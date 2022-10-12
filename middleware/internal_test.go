package middleware

import (
	"context"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"
)

func TestHookMiddleware(t *testing.T) {
	a := require.New(t)
	testErr := errors.New("test error")

	req := Request{
		Context: context.Background(),
		Body:    struct{}{},
	}
	_, err := HookMiddleware(
		func(req Request, next Next) (Response, error) {
			return next(req)
		},
		req,
		nil,
		func(context.Context, struct{}, struct{}) (r struct{}, _ error) {
			return struct{}{}, testErr
		},
	)
	a.ErrorIs(err, testErr)

	_, err = HookMiddleware(
		func(req Request, next Next) (r Response, _ error) {
			return r, testErr
		},
		req,
		nil,
		func(context.Context, struct{}, struct{}) (r struct{}, _ error) {
			a.Fail("Should not be called")
			return r, nil
		},
	)
	a.ErrorIs(err, testErr)
}
