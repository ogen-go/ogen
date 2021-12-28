package parser

import (
	"reflect"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/oas"
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
		Contents: make(map[string]*oas.Schema, len(body.Content)),
		Required: body.Required,
	}

	for contentType, media := range body.Content {
		if reflect.DeepEqual(media.Schema, ogen.Schema{}) {
			switch contentType {
			case "application/octet-stream":
				result.Contents[contentType] = nil
				continue
			default:
			}
		}

		schema, err := p.parseSchema(&media.Schema)
		if err != nil {
			return nil, errors.Wrapf(err, "content: %s: parse schema", contentType)
		}

		result.Contents[contentType] = schema
	}

	return result, nil
}
