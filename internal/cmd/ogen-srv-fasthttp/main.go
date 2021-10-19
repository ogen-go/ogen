package main

import (
	"context"
	"flag"
	"log"
	_ "net/http/pprof"

	"github.com/valyala/fasthttp"

	"github.com/ogen-go/ogen/internal/techempower"
	"github.com/ogen-go/ogen/json"
)

type server struct {
}

func (s server) Caching(ctx context.Context, params techempower.CachingParams) ([]techempower.WorldObject, error) {
	panic("implement me")
}

func (s server) DB(ctx context.Context) (techempower.WorldObject, error) {
	panic("implement me")
}

func (s server) JSON(ctx context.Context) (techempower.HelloWorld, error) {
	return techempower.HelloWorld{Message: "Hello, world"}, nil
}

func (s server) Queries(ctx context.Context, params techempower.QueriesParams) ([]techempower.WorldObject, error) {
	panic("implement me")
}

func (s server) Updates(ctx context.Context, params techempower.UpdatesParams) ([]techempower.WorldObject, error) {
	panic("implement me")
}

func main() {
	var arg struct {
		Addr string
	}
	flag.StringVar(&arg.Addr, "addr", "localhost:8080", "addr to listen")
	flag.Parse()
	s := &server{}
	log.Fatal(fasthttp.ListenAndServe(arg.Addr, func(ctx *fasthttp.RequestCtx) {
		ctx.SetContentType("application/json")
		res, _ := s.JSON(context.Background())
		buf := json.GetBuffer()
		defer json.PutBuffer(buf)
		_ = res.WriteJSONTo(buf)
		ctx.Write(buf.Bytes())
	}))
}
