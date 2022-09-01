package parser

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/jsonpointer"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseComponents(c *ogen.Components) (_ *openapi.Components, rerr error) {
	if c == nil {
		return &openapi.Components{
			Schemas:       map[string]*jsonschema.Schema{},
			Responses:     map[string]*openapi.Response{},
			Parameters:    map[string]*openapi.Parameter{},
			Examples:      map[string]*openapi.Example{},
			RequestBodies: map[string]*openapi.RequestBody{},
		}, nil
	}
	defer func() {
		rerr = p.wrapLocation("", c.Locator, rerr)
	}()

	result := &openapi.Components{
		Schemas:       make(map[string]*jsonschema.Schema, len(c.Schemas)),
		Responses:     make(map[string]*openapi.Response, len(c.Responses)),
		Parameters:    make(map[string]*openapi.Parameter, len(c.Parameters)),
		Examples:      make(map[string]*openapi.Example, len(c.Examples)),
		RequestBodies: make(map[string]*openapi.RequestBody, len(c.RequestBodies)),
	}
	wrapErr := func(component, name string, err error) error {
		loc := c.Locator.Field(component).Field(name)
		err = errors.Wrapf(err, "schemas: %q", name)
		return p.wrapLocation("", loc, err)
	}

	for name := range c.Schemas {
		ref := "#/components/schemas/" + name
		s, err := p.schemaParser.Resolve(ref, jsonpointer.NewResolveCtx(p.depthLimit))
		if err != nil {
			return nil, wrapErr("schemas", name, err)
		}

		result.Schemas[name] = s
	}

	for name := range c.Responses {
		ref := "#/components/responses/" + name
		r, err := p.resolveResponse(ref, jsonpointer.NewResolveCtx(p.depthLimit))
		if err != nil {
			return nil, wrapErr("responses", name, err)
		}

		result.Responses[name] = r
	}

	for name := range c.Parameters {
		ref := "#/components/parameters/" + name
		pp, err := p.resolveParameter(ref, jsonpointer.NewResolveCtx(p.depthLimit))
		if err != nil {
			return nil, wrapErr("parameters", name, err)
		}

		result.Parameters[name] = pp
	}

	for name := range c.Examples {
		ref := "#/components/examples/" + name
		ex, err := p.resolveExample(ref, jsonpointer.NewResolveCtx(p.depthLimit))
		if err != nil {
			return nil, wrapErr("examples", name, err)
		}

		result.Examples[name] = ex
	}

	for name := range c.RequestBodies {
		ref := "#/components/requestBodies/" + name
		b, err := p.resolveRequestBody(ref, jsonpointer.NewResolveCtx(p.depthLimit))
		if err != nil {
			return nil, wrapErr("requestBodies", name, err)
		}

		result.RequestBodies[name] = b
	}

	return result, nil
}
