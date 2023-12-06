package parser

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonpointer"
	"github.com/ogen-go/ogen/location"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseResponses(
	responses ogen.Responses,
	locator location.Locator,
	ctx *jsonpointer.ResolveCtx,
) (result openapi.Responses, _ error) {
	if len(responses) == 0 {
		return result, errors.New("no responses")
	}

	result = openapi.Responses{
		Pointer: locator.Pointer(p.file(ctx)),
	}
	for pattern, response := range responses {
		resp, err := p.parseResponse(response, ctx)
		if err != nil {
			err := errors.Wrap(err, pattern)
			return result, p.wrapLocation(p.file(ctx), locator.Field(pattern), err)
		}

		if err := result.Add(pattern, resp); err != nil {
			return result, p.wrapLocation(p.file(ctx), locator.Key(pattern), err)
		}
	}

	return result, nil
}

func (p *parser) parseResponse(resp *ogen.Response, ctx *jsonpointer.ResolveCtx) (_ *openapi.Response, rerr error) {
	if resp == nil {
		return nil, errors.New("response object is empty or null")
	}
	locator := resp.Common.Locator
	defer func() {
		rerr = p.wrapLocation(p.file(ctx), locator, rerr)
	}()
	if ref := resp.Ref; ref != "" {
		resolved, err := p.resolveResponse(ref, ctx)
		if err != nil {
			return nil, p.wrapRef(p.file(ctx), locator, err)
		}
		return resolved, nil
	}

	content, err := p.parseContent(resp.Content, locator.Field("content"), ctx)
	if err != nil {
		err := errors.Wrap(err, "content")
		return nil, p.wrapField("content", p.file(ctx), locator, err)
	}

	headers, err := p.parseHeaders(resp.Headers, ctx)
	if err != nil {
		err := errors.Wrap(err, "headers")
		return nil, p.wrapField("headers", p.file(ctx), locator, err)
	}

	return &openapi.Response{
		Description: resp.Description,
		Headers:     headers,
		Content:     content,
		Pointer:     locator.Pointer(p.file(ctx)),
	}, nil
}
