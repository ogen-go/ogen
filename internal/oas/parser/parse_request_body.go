package parser

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/oas"
	"github.com/ogen-go/ogen/jsonschema"
)

func (p *parser) parseRequestBody(body *ogen.RequestBody) (*oas.RequestBody, error) {
	if ref := body.Ref; ref != "" {
		reqBody, err := p.resolveRequestBody(ref)
		if err != nil {
			return nil, errors.Wrapf(err, "resolve %q reference", ref)
		}

		return reqBody, nil
	}

	result := &oas.RequestBody{
		Contents: make(map[string]*jsonschema.Schema, len(body.Content)),
		Required: body.Required,
	}

	for contentType, media := range body.Content {
		schema, err := p.schemaParser.Parse(media.Schema.ToJSONSchema())
		if err != nil {
			return nil, errors.Wrapf(err, "content: %s: parse schema", contentType)
		}

		result.Contents[contentType] = schema
	}

	return result, nil
}
