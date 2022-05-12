package parser

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseRequestBody(body *ogen.RequestBody, ctx resolveCtx) (*openapi.RequestBody, error) {
	if body == nil {
		return nil, errors.New("requestBody object is empty or null")
	}
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

	content, err := p.parseContent(body.Content)
	if err != nil {
		return nil, errors.Wrap(err, "content")
	}

	return &openapi.RequestBody{
		Content:  content,
		Required: body.Required,
	}, nil
}
