package parser

import (
	"regexp"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/jsonpointer"
	"github.com/ogen-go/ogen/internal/location"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
)

var componentsKeyRegex = regexp.MustCompile(`^[a-zA-Z0-9.\-_]+$`)

// validateComponentsKey validates components key.
//
// Spec says:
//
//	All the fixed fields declared above are objects that MUST use keys that
//	match the regular expression: ^[a-zA-Z0-9\.\-_]+$.
//
// See https://spec.openapis.org/oas/v3.1.0#components-object.
func validateComponentsKey[Object any](p *parser, m map[string]Object, locator location.Locator) error {
	for name := range m {
		if !componentsKeyRegex.MatchString(name) {
			locator := locator.Key(name)
			err := errors.Errorf("invalid name: %q doesn't match %q", name, componentsKeyRegex)
			return p.wrapLocation("", locator, err)
		}
	}
	return nil
}

// validateComponentsKeys validates components keys.
//
// See validateComponentsKey comment.
func validateComponentsKeys(p *parser, c *ogen.Components) error {
	if c == nil {
		return nil
	}
	if err := validateComponentsKey(p, c.Schemas, c.Locator.Field("schemas")); err != nil {
		return errors.Wrap(err, "schemas")
	}
	if err := validateComponentsKey(p, c.Responses, c.Locator.Field("responses")); err != nil {
		return errors.Wrap(err, "responses")
	}
	if err := validateComponentsKey(p, c.Parameters, c.Locator.Field("parameters")); err != nil {
		return errors.Wrap(err, "parameters")
	}
	if err := validateComponentsKey(p, c.Examples, c.Locator.Field("examples")); err != nil {
		return errors.Wrap(err, "examples")
	}
	if err := validateComponentsKey(p, c.RequestBodies, c.Locator.Field("requestBodies")); err != nil {
		return errors.Wrap(err, "requestBodies")
	}
	if err := validateComponentsKey(p, c.Headers, c.Locator.Field("headers")); err != nil {
		return errors.Wrap(err, "headers")
	}
	if err := validateComponentsKey(p, c.SecuritySchemes, c.Locator.Field("securitySchemes")); err != nil {
		return errors.Wrap(err, "securitySchemes")
	}
	if err := validateComponentsKey(p, c.Links, c.Locator.Field("links")); err != nil {
		return errors.Wrap(err, "links")
	}
	if err := validateComponentsKey(p, c.Callbacks, c.Locator.Field("callbacks")); err != nil {
		return errors.Wrap(err, "callbacks")
	}
	if err := validateComponentsKey(p, c.PathItems, c.Locator.Field("pathItems")); err != nil {
		return errors.Wrap(err, "pathItems")
	}
	return nil
}

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

	if err := validateComponentsKeys(p, c); err != nil {
		return nil, err
	}

	result := &openapi.Components{
		Schemas:       make(map[string]*jsonschema.Schema, len(c.Schemas)),
		Responses:     make(map[string]*openapi.Response, len(c.Responses)),
		Parameters:    make(map[string]*openapi.Parameter, len(c.Parameters)),
		Examples:      make(map[string]*openapi.Example, len(c.Examples)),
		RequestBodies: make(map[string]*openapi.RequestBody, len(c.RequestBodies)),
	}
	wrapErr := func(component, name string, err error) error {
		loc := c.Locator.Field(component).Field(name)
		err = errors.Wrapf(err, "%s: %q", component, name)
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
