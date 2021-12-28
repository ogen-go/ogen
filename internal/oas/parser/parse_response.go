package parser

import (
	"net/http"
	"reflect"
	"strconv"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/oas"
)

func (p *parser) parseResponses(responses ogen.Responses) (map[string]*oas.Response, error) {
	result := make(map[string]*oas.Response, len(responses))
	if len(responses) == 0 {
		return nil, errors.New("no responses")
	}

	for status, response := range responses {
		if err := validateStatusCode(status); err != nil {
			return nil, errors.Wrap(err, status)
		}

		resp, err := p.parseResponse(response)
		if err != nil {
			return nil, errors.Wrap(err, status)
		}

		result[status] = resp
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

	response := &oas.Response{
		Contents: make(map[string]*oas.Schema, len(resp.Content)),
	}
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

func validateStatusCode(v string) error {
	switch v {
	case "default", "1XX", "2XX", "3XX", "4XX", "5XX":
		return nil

	default:
		code, err := strconv.Atoi(v)
		if err != nil {
			return errors.Wrap(err, "parse status code")
		}

		if http.StatusText(code) == "" {
			return errors.Errorf("unknown status code: %d", code)
		}

		return nil
	}
}
