package jsonschema

import (
	"strings"

	"github.com/go-faster/errors"
	"github.com/go-faster/yaml"

	"github.com/ogen-go/ogen/jsonpointer"
)

// RootResolver is ReferenceResolver implementation.
type RootResolver struct {
	root *yaml.Node
}

// NewRootResolver creates new RootResolver.
func NewRootResolver(root *yaml.Node) *RootResolver {
	return &RootResolver{root: root}
}

// ResolveReference implements ReferenceResolver.
func (r *RootResolver) ResolveReference(ref string) (rawSchema *RawSchema, err error) {
	ref = strings.TrimSpace(ref)

	n, err := jsonpointer.Resolve(ref, r.root)
	if err != nil {
		return nil, errors.Wrap(err, "resolve")
	}

	rawSchema = &RawSchema{}
	if err := n.Decode(rawSchema); err != nil {
		return nil, errors.Wrap(err, "decode")
	}

	return rawSchema, nil
}
