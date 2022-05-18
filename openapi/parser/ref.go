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
		return nil, errors.New("invalid requestBody reference")
	}

	if r, ok := p.refs.requestBodies[ref]; ok {
		return r, nil
	}

	if _, ok := ctx[ref]; ok {
		return nil, errors.New("infinite recursion")
	}
	ctx[ref] = struct{}{}

	name := strings.TrimPrefix(ref, prefix)
	component, found := p.spec.Components.RequestBodies[name]
	if !found {
		return nil, errors.New("component by reference not found")
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
		return nil, errors.New("invalid response reference")
	}

	if r, ok := p.refs.responses[ref]; ok {
		return r, nil
	}

	if _, ok := ctx[ref]; ok {
		return nil, errors.New("infinite recursion")
	}
	ctx[ref] = struct{}{}

	name := strings.TrimPrefix(ref, prefix)
	component, found := p.spec.Components.Responses[name]
	if !found {
		return nil, errors.New("component by reference not found")
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
		return nil, errors.New("invalid parameter reference")
	}

	if param, ok := p.refs.parameters[ref]; ok {
		return param, nil
	}

	if _, ok := ctx[ref]; ok {
		return nil, errors.New("infinite recursion")
	}
	ctx[ref] = struct{}{}

	name := strings.TrimPrefix(ref, prefix)
	component, found := p.spec.Components.Parameters[name]
	if !found {
		return nil, errors.New("component by reference not found")
	}

	param, err := p.parseParameter(component, ctx)
	if err != nil {
		return nil, err
	}

	param.Ref = ref
	p.refs.parameters[ref] = param
	return param, nil
}

func (p *parser) resolveHeader(headerName, ref string, ctx resolveCtx) (*openapi.Header, error) {
	const prefix = "#/components/headers/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, errors.New("invalid header reference")
	}

	if param, ok := p.refs.headers[ref]; ok {
		return param, nil
	}

	if _, ok := ctx[ref]; ok {
		return nil, errors.New("infinite recursion")
	}
	ctx[ref] = struct{}{}

	name := strings.TrimPrefix(ref, prefix)
	component, found := p.spec.Components.Headers[name]
	if !found {
		return nil, errors.New("component by reference not found")
	}

	param, err := p.parseHeader(headerName, component, ctx)
	if err != nil {
		return nil, err
	}

	param.Ref = ref
	p.refs.headers[ref] = param
	return param, nil
}

func (p *parser) resolveExample(ref string, ctx resolveCtx) (*openapi.Example, error) {
	const prefix = "#/components/examples/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, errors.New("invalid example reference")
	}

	if param, ok := p.refs.examples[ref]; ok {
		return param, nil
	}

	if _, ok := ctx[ref]; ok {
		return nil, errors.New("infinite recursion")
	}
	ctx[ref] = struct{}{}

	name := strings.TrimPrefix(ref, prefix)
	component, found := p.spec.Components.Examples[name]
	if !found {
		return nil, errors.New("component by reference not found")
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
		return nil, errors.New("invalid securitySchema reference")
	}

	if param, ok := p.refs.securitySchemes[ref]; ok {
		return param, nil
	}

	if _, ok := ctx[ref]; ok {
		return nil, errors.New("infinite recursion")
	}
	ctx[ref] = struct{}{}

	name := strings.TrimPrefix(ref, prefix)
	component, found := p.spec.Components.SecuritySchemes[name]
	if !found {
		return nil, errors.New("component by reference not found")
	}

	p.refs.securitySchemes[ref] = component
	return component, nil
}
