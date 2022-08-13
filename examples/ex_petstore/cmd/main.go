package main

import (
	"context"
	"log"
	"net/http"

	"github.com/davecgh/go-spew/spew"
	"n.com/pets/api"
	"n.com/pets/model"
)

type Handler struct{}

func (h *Handler) ListPets(ctx context.Context, params api.ListPetsParams) (api.ListPetsRes, error) {
	m := model.NewModel()
	spew.Dump(params)

	pets := make(api.Pets, 0)
	resP := m.GetPets(params)

	for _, p := range resP {
		resP := api.Pet{ID: p.ID, Name: p.Name}
		pets = append(pets, resP)
	}

	res := api.PetsHeaders{
		Response: pets,
	}

	return &res, nil
}

func main() {
	h := &Handler{}

	server, err := api.NewServer(h)
	if err != nil {
		panic(err)
	}

	log.Fatal(http.ListenAndServe(":8080", server))
}
