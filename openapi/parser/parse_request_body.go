package parser

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseRequestBody(body *ogen.RequestBody, ctx *resolveCtx) (_ *openapi.RequestBody, rerr error) {
	if body == nil {
		return nil, errors.New("requestBody object is empty or null")
	}
	defer func() {
		rerr = p.wrapLocation(body, rerr)
	}()
	if ref := body.Ref; ref != "" {
		reqBody, err := p.resolveRequestBody(ref, ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "resolve %q reference", ref)
		}

		return reqBody, nil
	}

	if len(body.Content) < 1 {
		// See https://github.com/OAI/OpenAPI-Specification/discussions/2875.
		return nil, errors.New("content must have at least one entry")
	}

	content, err := p.parseContent(body.Content, ctx)
	if err != nil {
		return nil, errors.Wrap(err, "content")
	}

	return &openapi.RequestBody{
		Description: body.Description,
		Required:    body.Required,
		Content:     content,
		Locator:     body.Locator,
	}, nil
}
