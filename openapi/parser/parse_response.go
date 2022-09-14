package parser

import (
	"strconv"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/jsonpointer"
	"github.com/ogen-go/ogen/internal/location"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseResponses(
	responses ogen.Responses,
	locator location.Locator,
	ctx *jsonpointer.ResolveCtx,
) (_ map[string]*openapi.Response, err error) {
	if len(responses) == 0 {
		return nil, errors.New("no responses")
	}

	result := make(map[string]*openapi.Response, len(responses))
	for status, response := range responses {
		if err := validateStatusCode(status); err != nil {
			return nil, p.wrapLocation(ctx.LastLoc(), locator.Key(status), err)
		}

		resp, err := p.parseResponse(response, ctx)
		if err != nil {
			err := errors.Wrap(err, status)
			return nil, p.wrapLocation(ctx.LastLoc(), locator.Field(status), err)
		}

		result[status] = resp
	}

	return result, nil
}

func (p *parser) parseResponse(resp *ogen.Response, ctx *jsonpointer.ResolveCtx) (_ *openapi.Response, rerr error) {
	if resp == nil {
		return nil, errors.New("response object is empty or null")
	}
	locator := resp.Common.Locator
	defer func() {
		rerr = p.wrapLocation(ctx.LastLoc(), locator, rerr)
	}()
	if ref := resp.Ref; ref != "" {
		resolved, err := p.resolveResponse(ref, ctx)
		if err != nil {
			return nil, p.wrapRef(ctx.LastLoc(), locator, err)
		}
		return resolved, nil
	}

	content, err := p.parseContent(resp.Content, locator.Field("content"), ctx)
	if err != nil {
		err := errors.Wrap(err, "content")
		return nil, p.wrapField("content", ctx.LastLoc(), locator, err)
	}

	headers, err := p.parseHeaders(resp.Headers, ctx)
	if err != nil {
		err := errors.Wrap(err, "headers")
		return nil, p.wrapField("headers", ctx.LastLoc(), locator, err)
	}

	return &openapi.Response{
		Description: resp.Description,
		Headers:     headers,
		Content:     content,
		Locator:     locator,
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
