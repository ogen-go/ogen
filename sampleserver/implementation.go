package sampleserver

import (
	"context"
)

type MockServer struct {
}

func (m MockServer) PetsGet(ctx context.Context) ([]Pet, error) {
	return []Pet{
		{Name: "Sobeka", ID: 5014},
	}, nil
}
