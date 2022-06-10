package jsonschema

import (
	"context"

	"github.com/go-faster/errors"
)

// ExternalResolver resolves external links.
type ExternalResolver interface {
	Get(ctx context.Context, loc string) ([]byte, error)
}

var _ ExternalResolver = NoExternal{}

// NoExternal is ExternalResolver that always returns error.
type NoExternal struct{}

func (n NoExternal) Get(context.Context, string) ([]byte, error) {
	return nil, errors.New("external references are disabled")
}
