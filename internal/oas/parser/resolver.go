package parser

import (
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonschema"
)

type componentsResolver struct {
	components map[string]*ogen.Schema
	root       *jsonschema.RootResolver
}

func (c componentsResolver) ResolveReference(ref string) (*jsonschema.RawSchema, error) {
	const prefix = "#/components/schemas/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, errors.New("invalid schema reference")
	}

	name := strings.TrimPrefix(ref, prefix)
	s, ok := c.components[name]
	if !ok {
		return c.root.ResolveReference(ref)
	}

	return s.ToJSONSchema(), nil
}
