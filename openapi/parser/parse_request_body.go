package parser

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/jsonpointer"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseRequestBody(body *ogen.RequestBody, ctx *jsonpointer.ResolveCtx) (_ *openapi.RequestBody, rerr error) {
	if body == nil {
		return nil, errors.New("requestBody object is empty or null")
	}
	defer func() {
		rerr = p.wrapLocation(ctx.LastLoc(), body.Locator, rerr)
	}()
	if ref := body.Ref; ref != "" {
		resolved, err := p.resolveRequestBody(ref, ctx)
		if err != nil {
			return nil, p.wrapRef(ctx.LastLoc(), body.Locator, err)
		}
		return resolved, nil
	}

	content, err := func() (map[string]*openapi.MediaType, error) {
		if len(body.Content) < 1 {
			// See https://github.com/OAI/OpenAPI-Specification/discussions/2875.
			return nil, errors.New("content must have at least one entry")
		}

		content, err := p.parseContent(body.Content, body.Locator.Field("content"), ctx)
		if err != nil {
			return nil, errors.Wrap(err, "parse content")
		}

		return content, nil
	}()
	if err != nil {
		return nil, p.wrapField("content", ctx.LastLoc(), body.Locator, err)
	}

	return &openapi.RequestBody{
		Description: body.Description,
		Required:    body.Required,
		Content:     content,
		Locator:     body.Locator,
	}, nil
}
