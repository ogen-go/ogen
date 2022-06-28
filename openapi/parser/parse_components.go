package parser

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseComponents(c *ogen.Components) (*openapi.Components, error) {
	result := &openapi.Components{
		Parameters:    make(map[string]*openapi.Parameter, len(c.Parameters)),
		Schemas:       make(map[string]*jsonschema.Schema, len(c.Schemas)),
		RequestBodies: make(map[string]*openapi.RequestBody, len(c.RequestBodies)),
		Responses:     make(map[string]*openapi.Response, len(c.Responses)),
	}

	for name := range c.Parameters {
		ref := "#/components/parameters/" + name
		pp, err := p.resolveParameter(ref, newResolveCtx(p.depthLimit))
		if err != nil {
			return nil, errors.Wrapf(err, "parameters: %q", name)
		}

		result.Parameters[name] = pp
	}

	for name := range c.Schemas {
		ref := "#/components/schemas/" + name
		s, err := p.schemaParser.Resolve(ref)
		if err != nil {
			return nil, errors.Wrapf(err, "schemas: %q", name)
		}

		result.Schemas[name] = s
	}

	for name := range c.RequestBodies {
		ref := "#/components/requestBodies/" + name
		b, err := p.resolveRequestBody(ref, newResolveCtx(p.depthLimit))
		if err != nil {
			return nil, errors.Wrapf(err, "requestBodies: %q", name)
		}

		result.RequestBodies[name] = b
	}

	for name := range c.Responses {
		ref := "#/components/responses/" + name
		r, err := p.resolveResponse(ref, newResolveCtx(p.depthLimit))
		if err != nil {
			return nil, errors.Wrapf(err, "responses: %q", name)
		}

		result.Responses[name] = r
	}

	return result, nil
}
