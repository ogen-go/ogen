package parser

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseHeaders(headers map[string]*ogen.Header) (_ map[string]*openapi.Header, err error) {
	if len(headers) == 0 {
		return nil, nil
	}

	result := make(map[string]*openapi.Header, len(headers))
	for name, m := range headers {
		result[name], err = p.parseHeader(name, m, resolveCtx{})
		if err != nil {
			return nil, errors.Wrap(err, name)
		}
	}

	return result, nil
}

func (p *parser) parseHeader(name string, header *ogen.Header, ctx resolveCtx) (*openapi.Header, error) {
	if header == nil {
		return nil, errors.New("header object is empty or null")
	}
	if ref := header.Ref; ref != "" {
		parsed, err := p.resolveHeader(name, ref, ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "resolve %q reference", ref)
		}
		return parsed, nil
	}

	if header.In != "" {
		return nil, errors.Errorf(`"in" MUST NOT be specified, got %q`, header.In)
	}
	if header.Name != "" {
		return nil, errors.Errorf(`"name" MUST NOT be specified, got %q`, header.Name)
	}

	if header.Schema != nil && header.Content != nil {
		return nil, errors.New("header MUST contain either a schema property, or a content property, but not both")
	}

	if header.Schema == nil && header.Content == nil {
		return nil, errors.New("header MUST contain either a schema property, or a content property")
	}

	if header.Content != nil && len(header.Content) < 1 {
		// https://github.com/OAI/OpenAPI-Specification/discussions/2875
		return nil, errors.New("content must have at least one entry")
	}

	schema, err := p.schemaParser.Parse(header.Schema.ToJSONSchema())
	if err != nil {
		return nil, errors.Wrap(err, "schema")
	}

	content, err := p.parseContent(header.Content)
	if err != nil {
		return nil, errors.Wrap(err, "content")
	}

	locatedIn := openapi.LocationHeader
	op := &openapi.Header{
		In:          locatedIn,
		Name:        name,
		Description: header.Description,
		Schema:      schema,
		Content:     content,
		Style:       inferParamStyle(locatedIn, header.Style),
		Explode:     inferParamExplode(locatedIn, header.Explode),
		Required:    header.Required,
	}

	if header.Content != nil {
		// TODO: Validate content?
		return op, nil
	}

	if err := validateParamStyle(op); err != nil {
		return nil, err
	}

	return op, nil
}
