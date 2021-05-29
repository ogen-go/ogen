package sampleserver

import (
	"context"
)

type Pet struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Server interface {
	// operationID
	PetsGet(ctx context.Context) ([]Pet, error)
}
