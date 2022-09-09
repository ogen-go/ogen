package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChainMiddlewares(t *testing.T) {
	type testKey struct{}
	chain := ChainMiddlewares(
		func(req Request, next Next) (Response, error) {
			s := req.Body.([]string)
			s = append(s, "first")
			req.Body = s

			return next(req)
		},
		func(req Request, next Next) (Response, error) {
			s := req.Body.([]string)
			s = append(s, "second")
			req.Body = s

			req.Params["second"] = "bar"
			return next(req)
		},
		func(req Request, next Next) (Response, error) {
			s := req.Body.([]string)
			s = append(s, "third")
			req.Body = s

			req.Context = context.WithValue(req.Context, testKey{}, "baz")
			return next(req)
		},
	)
	a := require.New(t)

	req := Request{
		Context: context.Background(),
		Body:    []string{},
		Params:  map[string]any{},
	}
	resp, err := chain(req, func(req Request) (Response, error) {
		a.Equal([]string{"first", "second", "third"}, req.Body)
		a.Equal("bar", req.Params["second"])
		a.Equal("baz", req.Context.Value(testKey{}))
		return Response{Type: "ok"}, nil
	})
	a.NoError(err)
	a.Equal("ok", resp.Type)
}
