package parser

import (
	"fmt"
	"strconv"

	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen"
	ast "github.com/ogen-go/ogen/internal/ast"
)

func (p *parser) parseResponses(responses ogen.Responses) (*ast.OperationResponse, error) {
	result := createAstOpResponse()
	if len(responses) == 0 {
		return nil, fmt.Errorf("no responses")
	}

	// Iterate over method responses...
	for status, response := range responses {
		// Default response.
		if status == "default" {
			resp, err := p.parseResponse(response)
			if err != nil {
				return nil, xerrors.Errorf("default: %w", err)
			}

			result.Default = resp
			continue
		}

		statusCode, err := strconv.Atoi(status)
		if err != nil {
			return nil, xerrors.Errorf("invalid status code: '%s'", status)
		}

		resp, err := p.parseResponse(response)
		if err != nil {
			return nil, xerrors.Errorf("%s: %w", status, err)
		}

		result.StatusCode[statusCode] = resp
	}

	return result, nil
}

func (p *parser) parseResponse(resp ogen.Response) (*ast.Response, error) {
	if ref := resp.Ref; ref != "" {
		resp, err := p.resolveResponse(ref)
		if err != nil {
			return nil, xerrors.Errorf("resolve '%s' reference: %w", ref, err)
		}

		return resp, nil
	}

	response := createAstResponse()
	for contentType, media := range resp.Content {
		schema, err := p.parseSchema(media.Schema)
		if err != nil {
			return nil, xerrors.Errorf("content: %s: schema: %w", contentType, err)
		}

		response.Contents[contentType] = schema
	}

	return response, nil
}
