package parser

import (
	"reflect"
	"strconv"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/oas"
)

func (p *parser) parseResponses(responses ogen.Responses) (*oas.OperationResponse, error) {
	result := createAstOpResponse()
	if len(responses) == 0 {
		return nil, errors.New("no responses")
	}

	// Iterate over method responses...
	for status, response := range responses {
		// Default response.
		if status == "default" {
			resp, err := p.parseResponse(response)
			if err != nil {
				return nil, errors.Wrap(err, "default")
			}

			result.Default = resp
			continue
		}

		statusCode, err := strconv.Atoi(status)
		if err != nil {
			return nil, errors.Errorf("invalid status code: %q", status)
		}

		resp, err := p.parseResponse(response)
		if err != nil {
			return nil, errors.Wrapf(err, "%s", status)
		}

		result.StatusCode[statusCode] = resp
	}

	return result, nil
}

func (p *parser) parseResponse(resp *ogen.Response) (*oas.Response, error) {
	if ref := resp.Ref; ref != "" {
		resp, err := p.resolveResponse(ref)
		if err != nil {
			return nil, errors.Wrapf(err, "resolve %q reference", ref)
		}

		return resp, nil
	}

	response := createAstResponse()
	for contentType, media := range resp.Content {
		if reflect.DeepEqual(media.Schema, ogen.Schema{}) {
			switch contentType {
			case "application/octet-stream":
				response.Contents[contentType] = nil
				continue
			default:
			}
		}

		schema, err := p.parseSchema(&media.Schema)
		if err != nil {
			return nil, errors.Wrapf(err, "content: %s: schema", contentType)
		}

		response.Contents[contentType] = schema
	}

	return response, nil
}
