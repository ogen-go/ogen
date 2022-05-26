package parser

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseContent(content map[string]ogen.Media) (_ map[string]*openapi.MediaType, err error) {
	if len(content) == 0 {
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

	encodings := make(map[string]*openapi.Encoding, len(m.Encoding))
	if s != nil && len(m.Encoding) > 0 {
		names := make(map[string]jsonschema.Property, len(s.Properties))
		for _, prop := range s.Properties {
			names[prop.Name] = prop
		}

		parseEncoding := func(name string, e ogen.Encoding) error {
			prop, ok := names[name]
			if !ok {
				return errors.New("unknown prop")
			}

			encoding := &openapi.Encoding{
				ContentType:   e.ContentType,
				Headers:       map[string]*openapi.Header{},
				Style:         inferParamStyle(openapi.LocationQuery, e.Style),
				Explode:       inferParamExplode(openapi.LocationQuery, e.Explode),
				AllowReserved: e.AllowReserved,
			}
			encoding.Headers, err = p.parseHeaders(e.Headers)
			if err != nil {
				return errors.Wrap(err, "headers")
			}
			encodings[name] = encoding

			if err := validateParamStyle(&openapi.Parameter{
				Name:     name,
				Schema:   prop.Schema,
				Content:  nil,
				In:       openapi.LocationQuery,
				Style:    encoding.Style,
				Explode:  encoding.Explode,
				Required: prop.Required,
			}); err != nil {
				return errors.Wrap(err, "param style")
			}

			return nil
		}

		for name, e := range m.Encoding {
			if err := parseEncoding(name, e); err != nil {
				return nil, errors.Wrapf(err, "encoding property %q", name)
			}
		}
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
		Encoding: encodings,
	}, nil
}
