package parser

import (
	"strconv"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseStatus(status string, response *ogen.Response) (*openapi.Response, error) {
	if err := validateStatusCode(status); err != nil {
		return nil, err
	}

	if response == nil {
		return nil, errors.New("response object is empty or null")
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
	if ref := resp.Ref; ref != "" {
		resp, err := p.resolveResponse(ref, ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "resolve %q reference", ref)
		}

		return resp, nil
	}

	response := &openapi.Response{
		Contents: make(map[string]*jsonschema.Schema, len(resp.Content)),
	}
	for contentType, media := range resp.Content {
		schema, err := p.schemaParser.Parse(media.Schema.ToJSONSchema())
		if err != nil {
			return nil, errors.Wrapf(err, "content: %q: schema", contentType)
		}
		schema.AddExample(media.Example)
		for _, example := range media.Examples {
			schema.AddExample(example.Value)
			if ref := example.Ref; ref != "" {
				r, err := p.resolveExample(ref)
				if err != nil {
					return nil, errors.Wrapf(err, "resolve: %q", ref)
				}
				schema.AddExample(r.Value)
			}
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

		if code < 100 || code > 599 {
			return errors.Errorf("unknown status code: %d", code)
		}
		return nil
	}
}
