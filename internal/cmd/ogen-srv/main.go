package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/go-chi/chi/v5"

	"github.com/ogen-go/ogen/internal/techempower"
)

type server struct {
}

func (s server) Caching(ctx context.Context, params techempower.CachingParams) (techempower.WorldObjects, error) {
	panic("implement me")
}

func (s server) DB(ctx context.Context) (techempower.WorldObject, error) {
	panic("implement me")
}

func (s server) JSON(ctx context.Context) (techempower.HelloWorld, error) {
	return techempower.HelloWorld{Message: "Hello, world"}, nil
}

func (s server) Queries(ctx context.Context, params techempower.QueriesParams) (techempower.WorldObjects, error) {
	panic("implement me")
}

func (s server) Updates(ctx context.Context, params techempower.UpdatesParams) (techempower.WorldObjects, error) {
	panic("implement me")
}

func main() {
	var arg struct {
		Addr string
	}
	flag.StringVar(&arg.Addr, "addr", "localhost:8080", "addr to listen")
	flag.Parse()

	mux := chi.NewMux()
	techempower.Register(mux, &server{})
	fmt.Printf("http://%s/json\n", arg.Addr)

	http.Handle("/", mux)
	log.Fatal(http.ListenAndServe(arg.Addr, nil))
}
