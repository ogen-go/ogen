package parser

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/jsonpointer"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseHeaders(headers map[string]*ogen.Header, ctx *jsonpointer.ResolveCtx) (_ map[string]*openapi.Header, err error) {
	if len(headers) == 0 {
		return nil, nil
	}

	result := make(map[string]*openapi.Header, len(headers))
	for name, h := range headers {
		result[name], err = p.parseHeader(name, h, ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "header %q", name)
		}
	}

	return result, nil
}

func (p *parser) parseHeader(name string, header *ogen.Header, ctx *jsonpointer.ResolveCtx) (_ *openapi.Header, rerr error) {
	if header == nil {
		return nil, errors.New("header object is empty or null")
	}
	defer func() {
		rerr = p.wrapLocation(ctx.LastLoc(), header.Locator, rerr)
	}()
	if ref := header.Ref; ref != "" {
		resolved, err := p.resolveHeader(name, ref, ctx)
		if err != nil {
			return nil, p.wrapRef(ctx.LastLoc(), header.Locator, err)
		}
		return resolved, nil
	}

	mustNotSpecified := func(name, got string) error {
		if got == "" {
			return nil
		}
		err := errors.Errorf(`%q MUST NOT be specified, got %q`, name, got)
		return p.wrapField(name, ctx.LastLoc(), header.Locator, err)
	}
	if err := mustNotSpecified("in", header.In); err != nil {
		return nil, err
	}
	if err := mustNotSpecified("name", header.Name); err != nil {
		return nil, err
	}
	locatedIn := openapi.LocationHeader

	if err := p.validateParameter(name, locatedIn, header, ctx.LastLoc()); err != nil {
		return nil, err
	}

	schema, err := p.parseSchema(header.Schema, ctx)
	if err != nil {
		err := errors.Wrap(err, "schema")
		return nil, p.wrapField("schema", ctx.LastLoc(), header.Locator, err)
	}

	content, err := p.parseParameterContent(header.Content, header.Locator.Field("content"), ctx)
	if err != nil {
		err := errors.Wrap(err, "content")
		return nil, p.wrapField("content", ctx.LastLoc(), header.Locator, err)
	}

	op := &openapi.Header{
		Name:        name,
		Description: header.Description,
		Deprecated:  header.Deprecated,
		Schema:      schema,
		Content:     content,
		In:          locatedIn,
		Style:       inferParamStyle(locatedIn, header.Style),
		Explode:     inferParamExplode(locatedIn, header.Explode),
		Required:    header.Required,
		Locator:     header.Locator,
	}

	// TODO: Validate content?
	if header.Content == nil {
		if err := p.validateParamStyle(op, ctx.LastLoc()); err != nil {
			return nil, err
		}
	}

	return op, nil
}
