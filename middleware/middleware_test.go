package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/openapi"
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

			req.Params[ParameterKey{
				Name: "second",
				In:   openapi.LocationQuery,
			}] = "bar"
			return next(req)
		},
		func(req Request, next Next) (Response, error) {
			s := req.Body.([]string)
			s = append(s, "third")
			req.Body = s

			req.Context = context.WithValue(req.Context, testKey{}, "baz")
			return next(req)
		},
		func(req Request, next Next) (Response, error) {
			s := req.Body.([]string)
			s = append(s, "fourth")
			req.Body = s

			req.RawBody = []byte("qux")
			return next(req)
		},
	)
	a := require.New(t)

	for i := range [2]struct{}{} {
		req := Request{
			Context: context.Background(),
			Body:    []string{},
			Params: Parameters{
				{"call", openapi.LocationPath}: i,
			},
		}
		resp, err := chain(req, func(req Request) (Response, error) {
			a.Equal([]string{"first", "second", "third", "fourth"}, req.Body)
			a.Equal("bar", func() any {
				v, ok := req.Params.Query("second")
				a.True(ok)
				return v
			}())
			a.Equal("baz", req.Context.Value(testKey{}))
			a.Equal(i, func() any {
				v, ok := req.Params.Path("call")
				a.True(ok)
				return v
			}())
			a.Equal([]byte("qux"), req.RawBody)

			{
				_, ok := req.Params.Header("call")
				a.False(ok)
				_, ok = req.Params.Cookie("call")
				a.False(ok)
			}

			return Response{Type: "ok"}, nil
		})
		a.NoError(err)
		a.Equal("ok", resp.Type)
	}
}

func BenchmarkChainMiddlewares(b *testing.B) {
	const N = 20
	noop := func(req Request, next Next) (Response, error) {
		return next(req)
	}

	var (
		chain = ChainMiddlewares(func() (r []Middleware) {
			for i := 0; i < N; i++ {
				r = append(r, noop)
			}
			return r
		}()...)
		req = Request{
			Context: context.Background(),
			Body:    []string{},
			Params:  Parameters{},
		}
		resp = Response{Type: "ok"}
		next = func(req Request) (Response, error) {
			return resp, nil
		}
	)

	b.ReportAllocs()
	b.ResetTimer()

	var (
		sinkResp Response
		sinkErr  error
	)

	for i := 0; i < b.N; i++ {
		sinkResp, sinkErr = chain(req, next)
	}

	if sinkErr != nil {
		b.Fatal(sinkErr)
	}
	if sinkResp != resp {
		b.Fatalf("Expected %v, got %v", resp, sinkResp)
	}
}
