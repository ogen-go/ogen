package parser

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseHeaders(headers map[string]*ogen.Header, ctx *resolveCtx) (_ map[string]*openapi.Header, err error) {
	if len(headers) == 0 {
		return nil, nil
	}

	result := make(map[string]*openapi.Header, len(headers))
	for name, m := range headers {
		result[name], err = p.parseHeader(name, m, ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "header %q", name)
		}
	}

	return result, nil
}

func (p *parser) parseHeader(name string, header *ogen.Header, ctx *resolveCtx) (*openapi.Header, error) {
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
	locatedIn := openapi.LocationHeader

	if err := validateParameter(name, locatedIn, header); err != nil {
		return nil, err
	}

	schema, err := p.parseSchema(header.Schema, ctx)
	if err != nil {
		return nil, errors.Wrap(err, "schema")
	}

	content, err := p.parseParameterContent(header.Content, ctx)
	if err != nil {
		return nil, errors.Wrap(err, "content")
	}

	op := &openapi.Header{
		Name:        name,
		Description: header.Description,
		Schema:      schema,
		Content:     content,
		In:          locatedIn,
		Style:       inferParamStyle(locatedIn, header.Style),
		Explode:     inferParamExplode(locatedIn, header.Explode),
		Required:    header.Required,
		Deprecated:  header.Deprecated,
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
