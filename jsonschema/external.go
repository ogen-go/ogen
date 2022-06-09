package jsonschema

import (
	"context"

	"github.com/go-faster/errors"
)

// ExternalResolver resolves external links.
type ExternalResolver interface {
	Get(ctx context.Context, loc string) (ReferenceResolver, error)
}

// noExternal is ExternalResolver that always returns error.
type noExternal struct{}

func (n noExternal) Get(context.Context, string) (ReferenceResolver, error) {
	return nil, errors.New("external references are disabled")
}
