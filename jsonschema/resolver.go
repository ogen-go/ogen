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

func splitURI(u string) (base, fragment string) {
	hash := strings.IndexByte(u, '#')
	if hash == -1 {
		return u, "#"
	}
	f := u[hash:]
	if f == "#/" {
		f = "#"
	}
	return u[0:hash], f
}

// ResolveReference implements ReferenceResolver.
func (r RootResolver) ResolveReference(ref string) (rawSchema *RawSchema, err error) {
	ref = strings.TrimSpace(ref)

	base, fragment := splitURI(ref)
	if base != "" {
		return nil, errors.Errorf("external base %q is not supported", base)
	}

	buf, err := jsonpointer.Resolve(fragment, r.root)
	if err != nil {
		return nil, errors.Wrapf(err, "resolve %q", ref)
	}

	if err := json.Unmarshal(buf, &rawSchema); err != nil {
		return nil, errors.Wrap(err, "unmarshal")
	}
	return rawSchema, nil
}
