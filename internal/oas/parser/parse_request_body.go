package parser

import (
	"reflect"

	"github.com/ogen-go/errors"

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

	reqBody := createAstRBody()
	reqBody.Required = body.Required

	for contentType, media := range body.Content {
		if reflect.DeepEqual(media.Schema, ogen.Schema{}) {
			switch contentType {
			case "application/octet-stream":
				reqBody.Contents[contentType] = nil
				continue
			default:
			}
		}

		schema, err := p.parseSchema(media.Schema)
		if err != nil {
			return nil, errors.Wrapf(err, "content: %s: parse schema", contentType)
		}

		reqBody.Contents[contentType] = schema
	}

	return reqBody, nil
}
