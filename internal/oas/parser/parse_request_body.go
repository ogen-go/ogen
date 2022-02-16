package parser

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/oas"
	"github.com/ogen-go/ogen/jsonschema"
)

func (p *parser) parseRequestBody(body *ogen.RequestBody, ctx resolveCtx) (*oas.RequestBody, error) {
	if ref := body.Ref; ref != "" {
		reqBody, err := p.resolveRequestBody(ref, ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "resolve %q reference", ref)
		}

		return reqBody, nil
	}
	if len(body.Content) < 1 {
		// See https://github.com/OAI/OpenAPI-Specification/discussions/2875.
		return nil, errors.New("content must have at least on entry")
	}

	result := &oas.RequestBody{
		Contents: make(map[string]*jsonschema.Schema, len(body.Content)),
		Required: body.Required,
	}

	for contentType, media := range body.Content {
		schema, err := p.schemaParser.Parse(media.Schema.ToJSONSchema())
		if err != nil {
			return nil, errors.Wrapf(err, "content: %q: parse schema", contentType)
		}

		result.Contents[contentType] = schema
	}

	return result, nil
}
