package parser

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseComponents(c *ogen.Components) (_ *openapi.Components, rerr error) {
	result := &openapi.Components{
		Schemas:       make(map[string]*jsonschema.Schema, len(c.Schemas)),
		Responses:     make(map[string]*openapi.Response, len(c.Responses)),
		Parameters:    make(map[string]*openapi.Parameter, len(c.Parameters)),
		Examples:      make(map[string]*openapi.Example, len(c.Examples)),
		RequestBodies: make(map[string]*openapi.RequestBody, len(c.RequestBodies)),
	}
	if c != nil {
		defer func() {
			rerr = p.wrapLocation(&c.Locator, rerr)
		}()
	}

	for name := range c.Schemas {
		ref := "#/components/schemas/" + name
		s, err := p.schemaParser.Resolve(ref)
		if err != nil {
			return nil, errors.Wrapf(err, "schemas: %q", name)
		}

		result.Schemas[name] = s
	}

	for name := range c.Responses {
		ref := "#/components/responses/" + name
		r, err := p.resolveResponse(ref, newResolveCtx(p.depthLimit))
		if err != nil {
			return nil, errors.Wrapf(err, "responses: %q", name)
		}

		result.Responses[name] = r
	}

	for name := range c.Parameters {
		ref := "#/components/parameters/" + name
		pp, err := p.resolveParameter(ref, newResolveCtx(p.depthLimit))
		if err != nil {
			return nil, errors.Wrapf(err, "parameters: %q", name)
		}

		result.Parameters[name] = pp
	}

	for name := range c.Examples {
		ref := "#/components/examples/" + name
		ex, err := p.resolveExample(ref, newResolveCtx(p.depthLimit))
		if err != nil {
			return nil, errors.Wrapf(err, "examples: %q", name)
		}

		result.Examples[name] = ex
	}

	for name := range c.RequestBodies {
		ref := "#/components/requestBodies/" + name
		b, err := p.resolveRequestBody(ref, newResolveCtx(p.depthLimit))
		if err != nil {
			return nil, errors.Wrapf(err, "requestBodies: %q", name)
		}

		result.RequestBodies[name] = b
	}

	return result, nil
}
