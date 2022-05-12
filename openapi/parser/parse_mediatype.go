package parser

import (
	"github.com/go-faster/errors"
	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseMediaType(m ogen.Media) (*openapi.MediaType, error) {
	s, err := p.schemaParser.Parse(m.Schema.ToJSONSchema())
	if err != nil {
		return nil, errors.Wrap(err, "schema")
	}

	examples := make(map[string]*openapi.Example, len(m.Examples))
	for name, ex := range m.Examples {
		e, err := p.parseExample(ex, resolveCtx{})
		if err != nil {
			return nil, errors.Wrapf(err, "examples: %q", name)
		}

		examples[name] = e
	}

	return &openapi.MediaType{
		Schema:   s,
		Example:  m.Example,
		Examples: examples,
	}, nil
}
