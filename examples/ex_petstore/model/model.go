package model

import "n.com/pets/api"

type Pet struct {
	ID   int64
	Name string
	Tag  string
}

type IModel interface {
	GetPets(params api.ListPetsParams) []Pet
}

type Model struct {
	pets []Pet
}

func (m *Model) GetPets(params api.ListPetsParams) []Pet {
	return []Pet{
		{ID: 1, Name: "Gaf", Tag: "asdf"},
	}
}

func NewModel() IModel {
	return &Model{}
}
