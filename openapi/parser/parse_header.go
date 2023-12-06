package parser

import (
	"fmt"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonpointer"
	"github.com/ogen-go/ogen/location"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseHeaders(headers map[string]*ogen.Header, ctx *jsonpointer.ResolveCtx) (_ map[string]*openapi.Header, err error) {
	if len(headers) == 0 {
		return nil, nil
	}

	uniq := map[string]location.Pointer{}
	result := make(map[string]*openapi.Header, len(headers))
	for name, h := range headers {
		parsed, err := p.parseHeader(name, h, ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "header %q", name)
		}

		canonicalName := canonicalParamName(name, openapi.LocationHeader)
		ptr := parsed.Pointer
		if existingPtr, ok := uniq[canonicalName]; ok {
			me := new(location.MultiError)
			me.ReportPtr(existingPtr, fmt.Sprintf("duplicate header: %q", name))
			me.ReportPtr(ptr, "")
			return nil, me
		}
		uniq[canonicalName] = ptr

		result[name] = parsed
	}

	return result, nil
}

func (p *parser) parseHeader(name string, header *ogen.Header, ctx *jsonpointer.ResolveCtx) (_ *openapi.Header, rerr error) {
	if header == nil {
		return nil, errors.New("header object is empty or null")
	}
	locator := header.Common.Locator
	defer func() {
		rerr = p.wrapLocation(p.file(ctx), locator, rerr)
	}()
	if ref := header.Ref; ref != "" {
		resolved, err := p.resolveHeader(name, ref, ctx)
		if err != nil {
			return nil, p.wrapRef(p.file(ctx), locator, err)
		}
		return resolved, nil
	}

	mustNotSpecified := func(name, got string) error {
		if got == "" {
			return nil
		}
		err := errors.Errorf(`%q MUST NOT be specified, got %q`, name, got)
		return p.wrapField(name, p.file(ctx), locator, err)
	}
	if err := mustNotSpecified("in", header.In); err != nil {
		return nil, err
	}
	if err := mustNotSpecified("name", header.Name); err != nil {
		return nil, err
	}
	locatedIn := openapi.LocationHeader

	if err := p.validateParameter(name, locatedIn, header, p.file(ctx)); err != nil {
		return nil, err
	}

	schema, err := p.parseSchema(header.Schema, ctx)
	if err != nil {
		return nil, p.wrapField("schema", p.file(ctx), locator, err)
	}

	content, err := p.parseParameterContent(header.Content, locator.Field("content"), ctx)
	if err != nil {
		err := errors.Wrap(err, "content")
		return nil, p.wrapField("content", p.file(ctx), locator, err)
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
		Pointer:     locator.Pointer(p.file(ctx)),
	}

	if err := p.validateParamStyle(op, p.file(ctx)); err != nil {
		return nil, err
	}

	return op, nil
}
