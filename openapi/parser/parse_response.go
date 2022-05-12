package parser

import (
	"strconv"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseStatus(status string, response *ogen.Response) (*openapi.Response, error) {
	if err := validateStatusCode(status); err != nil {
		return nil, err
	}

	resp, err := p.parseResponse(response, resolveCtx{})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (p *parser) parseResponses(responses ogen.Responses) (map[string]*openapi.Response, error) {
	result := make(map[string]*openapi.Response, len(responses))
	if len(responses) == 0 {
		return nil, errors.New("no responses")
	}

	for status, response := range responses {
		resp, err := p.parseStatus(status, response)
		if err != nil {
			return nil, errors.Wrap(err, status)
		}
		result[status] = resp
	}

	return result, nil
}

func (p *parser) parseResponse(resp *ogen.Response, ctx resolveCtx) (*openapi.Response, error) {
	if resp == nil {
		return nil, errors.New("response object is empty or null")
	}
	if ref := resp.Ref; ref != "" {
		resp, err := p.resolveResponse(ref, ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "resolve %q reference", ref)
		}

		return resp, nil
	}

	content, err := p.parseContent(resp.Content)
	if err != nil {
		return nil, errors.Wrap(err, "content")
	}

	return &openapi.Response{
		Content: content,
	}, nil
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

		if code < 100 || code > 599 {
			return errors.Errorf("unknown status code: %d", code)
		}
		return nil
	}
}
