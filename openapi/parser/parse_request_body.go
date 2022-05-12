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

	result := &openapi.RequestBody{
		Content:  make(map[string]*openapi.MediaType, len(body.Content)),
		Required: body.Required,
	}

	for contentType, media := range body.Content {
		m, err := p.parseMediaType(media)
		if err != nil {
			return nil, errors.Wrapf(err, "content: %q", contentType)
		}

		result.Content[contentType] = m
	}

	return result, nil
}
