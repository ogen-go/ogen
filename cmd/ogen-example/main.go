package main

import (
	"context"
	"log"
	"net/http"

	"github.com/ernado/ogen/api"
	"github.com/go-chi/chi/v5"
)

type server struct{}

func (s server) PetGet(ctx context.Context) (*api.Pet, error) {
	return &api.Pet{
		ID:   1337,
		Name: "DOG",
	}, nil
}

func main() {
	mux := chi.NewRouter()
	api.Register(mux, server{})
	log.Fatal(http.ListenAndServe(":8080", mux))
}
