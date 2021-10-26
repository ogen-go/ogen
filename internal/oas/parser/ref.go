package parser

import (
	"fmt"
	"strings"

	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen/internal/oas"
)

func (p *parser) resolveRequestBody(ref string) (*oas.RequestBody, error) {
	const prefix = "#/components/requestBodies/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, xerrors.Errorf("invalid requestBody reference: '%s'", ref)
	}

	if r, ok := p.refs.requestBodies[ref]; ok {
		return r, nil
	}

	name := strings.TrimPrefix(ref, prefix)
	component, found := p.spec.Components.RequestBodies[name]
	if !found {
		return nil, fmt.Errorf("component by reference '%s' not found", ref)
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
		return nil, xerrors.Errorf("invalid response reference: '%s'", ref)
	}

	if r, ok := p.refs.responses[ref]; ok {
		return r, nil
	}

	name := strings.TrimPrefix(ref, prefix)
	component, found := p.spec.Components.Responses[name]
	if !found {
		return nil, fmt.Errorf("component by reference '%s' not found", ref)
	}

	r, err := p.parseResponse(component)
	if err != nil {
		return nil, err
	}

	p.refs.responses[ref] = r
	return r, nil
}

func (p *parser) resolveParameter(ref string) (*oas.Parameter, error) {
	const prefix = "#/components/parameters/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, xerrors.Errorf("invalid parameter reference: '%s'", ref)
	}

	if param, ok := p.refs.parameters[ref]; ok {
		return param, nil
	}

	name := strings.TrimPrefix(ref, prefix)
	component, found := p.spec.Components.Parameters[name]
	if !found {
		return nil, fmt.Errorf("component by reference '%s' not found", ref)
	}

	param, err := p.parseParameter(component)
	if err != nil {
		return nil, err
	}

	p.refs.parameters[ref] = param
	return param, nil
}
