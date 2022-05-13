package parser

import (
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/openapi"
)

type resolveCtx map[string]struct{}

func (p *parser) resolveRequestBody(ref string, ctx resolveCtx) (*openapi.RequestBody, error) {
	const prefix = "#/components/requestBodies/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, errors.Errorf("invalid requestBody reference: %q", ref)
	}

	if r, ok := p.refs.requestBodies[ref]; ok {
		return r, nil
	}

	if _, ok := ctx[ref]; ok {
		return nil, errors.Errorf("infinite recursion: %q", ref)
	}
	ctx[ref] = struct{}{}

	name := strings.TrimPrefix(ref, prefix)
	component, found := p.spec.Components.RequestBodies[name]
	if !found {
		return nil, errors.Errorf("component by reference %q not found", ref)
	}

	r, err := p.parseRequestBody(component, ctx)
	if err != nil {
		return nil, err
	}

	r.Ref = ref
	p.refs.requestBodies[ref] = r
	return r, nil
}

func (p *parser) resolveResponse(ref string, ctx resolveCtx) (*openapi.Response, error) {
	const prefix = "#/components/responses/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, errors.Errorf("invalid response reference: %q", ref)
	}

	if r, ok := p.refs.responses[ref]; ok {
		return r, nil
	}

	if _, ok := ctx[ref]; ok {
		return nil, errors.Errorf("infinite recursion: %q", ref)
	}
	ctx[ref] = struct{}{}

	name := strings.TrimPrefix(ref, prefix)
	component, found := p.spec.Components.Responses[name]
	if !found {
		return nil, errors.Errorf("component by reference %q not found", ref)
	}

	r, err := p.parseResponse(component, ctx)
	if err != nil {
		return nil, err
	}

	r.Ref = ref
	p.refs.responses[ref] = r
	return r, nil
}

func (p *parser) resolveParameter(ref string, ctx resolveCtx) (*openapi.Parameter, error) {
	const prefix = "#/components/parameters/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, errors.Errorf("invalid parameter reference: %q", ref)
	}

	if param, ok := p.refs.parameters[ref]; ok {
		return param, nil
	}

	if _, ok := ctx[ref]; ok {
		return nil, errors.Errorf("infinite recursion: %q", ref)
	}
	ctx[ref] = struct{}{}

	name := strings.TrimPrefix(ref, prefix)
	component, found := p.spec.Components.Parameters[name]
	if !found {
		return nil, errors.Errorf("component by reference %q not found", ref)
	}

	param, err := p.parseParameter(component, ctx)
	if err != nil {
		return nil, err
	}

	param.Ref = ref
	p.refs.parameters[ref] = param
	return param, nil
}

func (p *parser) resolveExample(ref string, ctx resolveCtx) (*openapi.Example, error) {
	const prefix = "#/components/examples/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, errors.Errorf("invalid example reference: %q", ref)
	}

	if param, ok := p.refs.examples[ref]; ok {
		return param, nil
	}

	if _, ok := ctx[ref]; ok {
		return nil, errors.Errorf("infinite recursion: %q", ref)
	}
	ctx[ref] = struct{}{}

	name := strings.TrimPrefix(ref, prefix)
	component, found := p.spec.Components.Examples[name]
	if !found {
		return nil, errors.Errorf("component by reference %q not found", ref)
	}

	ex, err := p.parseExample(component, ctx)
	if err != nil {
		return nil, err
	}

	ex.Ref = ref
	p.refs.examples[ref] = ex
	return ex, nil
}

func (p *parser) resolveSecuritySchema(ref string, ctx resolveCtx) (*ogen.SecuritySchema, error) {
	const prefix = "#/components/securitySchemes/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, errors.Errorf("invalid securitySchema reference: %q", ref)
	}

	if param, ok := p.refs.securitySchemes[ref]; ok {
		return param, nil
	}

	if _, ok := ctx[ref]; ok {
		return nil, errors.Errorf("infinite recursion: %q", ref)
	}
	ctx[ref] = struct{}{}

	name := strings.TrimPrefix(ref, prefix)
	component, found := p.spec.Components.SecuritySchemes[name]
	if !found {
		return nil, errors.Errorf("component by reference %q not found", ref)
	}

	p.refs.securitySchemes[ref] = component
	return component, nil
}
