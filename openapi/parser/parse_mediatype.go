package parser

import (
	"encoding/json"
	"mime"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseContent(content map[string]ogen.Media, ctx *resolveCtx) (_ map[string]*openapi.MediaType, err error) {
	if len(content) == 0 {
		return nil, nil
	}

	result := make(map[string]*openapi.MediaType, len(content))
	for name, m := range content {
		if _, _, err := mime.ParseMediaType(name); err != nil {
			return nil, errors.Wrapf(err, "content type %q", name)
		}

		result[name], err = p.parseMediaType(m, ctx)
		if err != nil {
			return nil, errors.Wrap(err, name)
		}
	}

	return result, nil
}

func (p *parser) parseParameterContent(content map[string]ogen.Media, ctx *resolveCtx) (*openapi.ParameterContent, error) {
	if content == nil {
		return nil, nil
	}
	if len(content) != 1 {
		return nil, errors.New(`"content" map MUST only contain one entry`)
	}

	for name, m := range content {
		if _, _, err := mime.ParseMediaType(name); err != nil {
			return nil, errors.Wrapf(err, "content type %q", name)
		}

		media, err := p.parseMediaType(m, ctx)
		if err != nil {
			return nil, errors.Wrap(err, name)
		}

		return &openapi.ParameterContent{
			Name:  name,
			Media: media,
		}, nil
	}
	panic("unreachable")
}

func (p *parser) parseMediaType(m ogen.Media, ctx *resolveCtx) (_ *openapi.MediaType, rerr error) {
	defer func() {
		rerr = p.wrapLocation(ctx.lastLoc(), m.Locator, rerr)
	}()

	s, err := p.parseSchema(m.Schema, ctx)
	if err != nil {
		return nil, errors.Wrap(err, "schema")
	}

	encodings := make(map[string]*openapi.Encoding, len(m.Encoding))
	if s != nil && len(m.Encoding) > 0 {
		names := make(map[string]jsonschema.Property, len(s.Properties))
		for _, prop := range s.Properties {
			names[prop.Name] = prop
		}

		parseEncoding := func(name string, e ogen.Encoding) (rerr error) {
			defer func() {
				rerr = p.wrapLocation(ctx.lastLoc(), e.Locator, rerr)
			}()

			prop, ok := names[name]
			if !ok {
				return errors.Errorf("unknown property %q", name)
			}

			encoding := &openapi.Encoding{
				ContentType:   e.ContentType,
				Headers:       map[string]*openapi.Header{},
				Style:         inferParamStyle(openapi.LocationQuery, e.Style),
				Explode:       inferParamExplode(openapi.LocationQuery, e.Explode),
				AllowReserved: e.AllowReserved,
				Locator:       e.Locator,
			}
			encoding.Headers, err = p.parseHeaders(e.Headers, ctx)
			if err != nil {
				return p.wrapField("headers", ctx.lastLoc(), e.Locator, err)
			}
			encodings[name] = encoding

			if err := p.validateParamStyle(&openapi.Parameter{
				Name:     name,
				Schema:   prop.Schema,
				In:       openapi.LocationQuery,
				Style:    encoding.Style,
				Explode:  encoding.Explode,
				Required: prop.Required,
				Locator:  encoding.Locator,
			}, ctx.lastLoc()); err != nil {
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
		examples[name], err = p.parseExample(ex, ctx)
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
		Example:  json.RawMessage(m.Example),
		Examples: examples,
		Encoding: encodings,
		Locator:  m.Locator,
	}, nil
}
