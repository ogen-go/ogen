package main

import (
	"context"
	"flag"
	"log"

	"github.com/go-chi/chi/v5"
	"github.com/valyala/fasthttp"

	"github.com/ogen-go/ogen/internal/techempower"
)

type server struct{}

func (server) JSON(ctx context.Context) (techempower.HelloWorld, error) {
	return techempower.HelloWorld{
		Message: "Hello, world",
	}, nil
}

func (server) DB(ctx context.Context) (techempower.WorldObject, error) { panic("implement me") }
func (server) Caching(ctx context.Context, params techempower.CachingParams) (techempower.WorldObjects, error) {
	panic("implement me")
}
func (server) Queries(ctx context.Context, params techempower.QueriesParams) (techempower.WorldObjects, error) {
	panic("implement me")
}
func (server) Updates(ctx context.Context, params techempower.UpdatesParams) (techempower.WorldObjects, error) {
	panic("implement me")
}

func main() {
	var arg struct {
		Addr string
	}
	flag.StringVar(&arg.Addr, "addr", ":8080", "http address to listen")
	flag.Parse()

	mux := chi.NewRouter()

	techempower.Register(mux, &server{})
	s := &server{}
	h := techempower.NewJSONFastHandler(s)
	log.Fatal(fasthttp.ListenAndServe(arg.Addr, h))
}
