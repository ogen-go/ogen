package parser

import (
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/oas"
)

func (p *parser) resolveRequestBody(ref string) (*oas.RequestBody, error) {
	const prefix = "#/components/requestBodies/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, errors.Errorf("invalid requestBody reference: %q", ref)
	}

	if r, ok := p.refs.requestBodies[ref]; ok {
		return r, nil
	}

	name := strings.TrimPrefix(ref, prefix)
	component, found := p.spec.Components.RequestBodies[name]
	if !found {
		return nil, errors.Errorf("component by reference %q not found", ref)
	}

	r, err := p.parseRequestBody(&component)
	if err != nil {
		return nil, err
	}

	p.refs.requestBodies[ref] = r
	return r, nil
}

func (p *parser) resolveResponse(ref string) (*oas.Response, error) {
	const prefix = "#/components/responses/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, errors.Errorf("invalid response reference: %q", ref)
	}

	if r, ok := p.refs.responses[ref]; ok {
		return r, nil
	}

	name := strings.TrimPrefix(ref, prefix)
	component, found := p.spec.Components.Responses[name]
	if !found {
		return nil, errors.Errorf("component by reference %q not found", ref)
	}

	r, err := p.parseResponse(component)
	if err != nil {
		return nil, err
	}

	r.Ref = ref
	p.refs.responses[ref] = r
	return r, nil
}

func (p *parser) resolveParameter(ref string) (*oas.Parameter, error) {
	const prefix = "#/components/parameters/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, errors.Errorf("invalid parameter reference: %q", ref)
	}

	if param, ok := p.refs.parameters[ref]; ok {
		return param, nil
	}

	name := strings.TrimPrefix(ref, prefix)
	component, found := p.spec.Components.Parameters[name]
	if !found {
		return nil, errors.Errorf("component by reference %q not found", ref)
	}

	param, err := p.parseParameter(component)
	if err != nil {
		return nil, err
	}

	p.refs.parameters[ref] = param
	return param, nil
}

func (p *parser) resolveSecurity(ref string) (*oas.Security, error) {
	const prefix = "#/components/securitySchemes/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, errors.Errorf("invalid security reference: %q", ref)
	}

	if security, ok := p.refs.security[ref]; ok {
		return security, nil
	}

	name := strings.TrimPrefix(ref, prefix)
	schema, found := p.spec.Components.SecuritySchemes[name]
	if !found {
		return nil, errors.Errorf("component by reference %q not found", ref)
	}

	security, err := p.parseSecurity(schema)
	if err != nil {
		return nil, err
	}

	p.refs.security[ref] = security
	return security, nil
}
