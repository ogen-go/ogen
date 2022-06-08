package jsonschema

import (
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/json"
	"github.com/ogen-go/ogen/jsonpointer"
)

// RootResolver is ReferenceResolver implementation.
type RootResolver struct {
	root []byte
}

// NewRootResolver creates new RootResolver.
func NewRootResolver(root []byte) *RootResolver {
	return &RootResolver{root: root}
}

// ResolveReference implements ReferenceResolver.
func (r RootResolver) ResolveReference(ref string) (rawSchema *RawSchema, err error) {
	ref = strings.TrimSpace(ref)

	buf, err := jsonpointer.Resolve(ref, r.root)
	if err != nil {
		return nil, errors.Wrapf(err, "resolve %q", ref)
	}

	if err := json.Unmarshal(buf, &rawSchema); err != nil {
		return nil, errors.Wrap(err, "unmarshal")
	}
	return rawSchema, nil
}
