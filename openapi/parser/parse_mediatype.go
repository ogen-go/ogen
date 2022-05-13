package parser

import (
	"github.com/go-faster/errors"
	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseContent(content map[string]ogen.Media) (_ map[string]*openapi.MediaType, err error) {
	if content == nil {
		return nil, nil
	}

	result := make(map[string]*openapi.MediaType, len(content))
	for name, m := range content {
		result[name], err = p.parseMediaType(m)
		if err != nil {
			return nil, errors.Wrap(err, name)
		}
	}

	return result, nil
}

func (p *parser) parseMediaType(m ogen.Media) (*openapi.MediaType, error) {
	s, err := p.schemaParser.Parse(m.Schema.ToJSONSchema())
	if err != nil {
		return nil, errors.Wrap(err, "schema")
	}

	examples := make(map[string]*openapi.Example, len(m.Examples))
	for name, ex := range m.Examples {
		examples[name], err = p.parseExample(ex, resolveCtx{})
		if err != nil {
			return nil, errors.Wrapf(err, "examples: %q", name)
		}
	}

	// OpenAPI 3.0.3 doc says:
	//
	//   Furthermore, referencing a schema which contains an example,
	//   the example value SHALL override the example provided by the schema.
	//
	// Probably this will be rewritten later.
	// Kept for backward compatibility.
	s.AddExample(m.Example)
	for _, ex := range examples {
		s.AddExample(ex.Value)

	}

	return &openapi.MediaType{
		Schema:   s,
		Example:  m.Example,
		Examples: examples,
	}, nil
}
