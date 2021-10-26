package parser

import (
	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen"
	ast "github.com/ogen-go/ogen/internal/oas"
)

func (p *parser) parseRequestBody(body *ogen.RequestBody) (*ast.RequestBody, error) {
	if ref := body.Ref; ref != "" {
		reqBody, err := p.resolveRequestBody(ref)
		if err != nil {
			return nil, xerrors.Errorf("resolve '%s' reference: %w", ref, err)
		}

		return reqBody, nil
	}

	reqBody := createAstRBody()
	reqBody.Required = body.Required

	// Iterate through request body contents...
	for contentType, media := range body.Content {
		schema, err := p.parseSchema(media.Schema)
		if err != nil {
			return nil, xerrors.Errorf("content: %s: parse schema: %w", contentType, err)
		}

		reqBody.Contents[contentType] = schema
	}

	return reqBody, nil
}
