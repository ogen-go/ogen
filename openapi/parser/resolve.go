package parser

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonpointer"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) getSchema(ctx *resolveCtx) ([]byte, error) {
	loc := ctx.lastLoc()

	r, ok := p.schemas[loc]
	if ok {
		return r, nil
	}

	r, err := p.external.Get(context.TODO(), loc)
	if err != nil {
		return nil, errors.Wrapf(err, "external %q", loc)
	}
	p.schemas[loc] = r

	return r, nil
}

func resolvePointer(root []byte, ptr string, to interface{}) error {
	data, err := jsonpointer.Resolve(ptr, root)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, to)
}

func (p *parser) resolve(key refKey, ctx *resolveCtx, to interface{}) error {
	schema, err := p.getSchema(ctx)
	if err != nil {
		return err
	}
	return resolvePointer(schema, key.ref, to)
}

func (p *parser) resolveRequestBody(ref string, ctx *resolveCtx) (*openapi.RequestBody, error) {
	const prefix = "#/components/requestBodies/"

	if r, ok := p.refs.requestBodies[ref]; ok {
		return r, nil
	}

	key, err := ctx.add(ref)
	if err != nil {
		return nil, err
	}
	defer func() {
		ctx.delete(key)
	}()

	var component *ogen.RequestBody
	if key.loc == "" && ctx.lastLoc() == "" {
		name := strings.TrimPrefix(ref, prefix)
		c, found := p.spec.Components.RequestBodies[name]
		if found {
			component = c
		} else {
			if err := resolvePointer(p.spec.Raw, ref, &component); err != nil {
				return nil, err
			}
		}
	} else {
		if err := p.resolve(key, ctx, &component); err != nil {
			return nil, err
		}
	}

	r, err := p.parseRequestBody(component, ctx)
	if err != nil {
		return nil, err
	}

	r.Ref = ref
	p.refs.requestBodies[ref] = r
	return r, nil
}

func (p *parser) resolveResponse(ref string, ctx *resolveCtx) (*openapi.Response, error) {
	const prefix = "#/components/responses/"

	if r, ok := p.refs.responses[ref]; ok {
		return r, nil
	}

	key, err := ctx.add(ref)
	if err != nil {
		return nil, err
	}
	defer func() {
		ctx.delete(key)
	}()

	var component *ogen.Response
	if key.loc == "" && ctx.lastLoc() == "" {
		name := strings.TrimPrefix(ref, prefix)
		c, found := p.spec.Components.Responses[name]
		if found {
			component = c
		} else {
			if err := resolvePointer(p.spec.Raw, ref, &component); err != nil {
				return nil, err
			}
		}
	} else {
		if err := p.resolve(key, ctx, &component); err != nil {
			return nil, err
		}
	}

	r, err := p.parseResponse(component, ctx)
	if err != nil {
		return nil, err
	}

	r.Ref = ref
	p.refs.responses[ref] = r
	return r, nil
}

func (p *parser) resolveParameter(ref string, ctx *resolveCtx) (*openapi.Parameter, error) {
	const prefix = "#/components/parameters/"

	if param, ok := p.refs.parameters[ref]; ok {
		return param, nil
	}

	key, err := ctx.add(ref)
	if err != nil {
		return nil, err
	}
	defer func() {
		ctx.delete(key)
	}()

	var component *ogen.Parameter
	if key.loc == "" && ctx.lastLoc() == "" {
		name := strings.TrimPrefix(ref, prefix)
		c, found := p.spec.Components.Parameters[name]
		if found {
			component = c
		} else {
			if err := resolvePointer(p.spec.Raw, ref, &component); err != nil {
				return nil, err
			}
		}
	} else {
		if err := p.resolve(key, ctx, &component); err != nil {
			return nil, err
		}
	}

	param, err := p.parseParameter(component, ctx)
	if err != nil {
		return nil, err
	}

	param.Ref = ref
	p.refs.parameters[ref] = param
	return param, nil
}

func (p *parser) resolveHeader(headerName, ref string, ctx *resolveCtx) (*openapi.Header, error) {
	const prefix = "#/components/headers/"

	if header, ok := p.refs.headers[ref]; ok {
		return header, nil
	}

	key, err := ctx.add(ref)
	if err != nil {
		return nil, err
	}
	defer func() {
		ctx.delete(key)
	}()

	var component *ogen.Header
	if key.loc == "" && ctx.lastLoc() == "" {
		name := strings.TrimPrefix(ref, prefix)
		c, found := p.spec.Components.Headers[name]
		if found {
			component = c
		} else {
			if err := resolvePointer(p.spec.Raw, ref, &component); err != nil {
				return nil, err
			}
		}
	} else {
		if err := p.resolve(key, ctx, &component); err != nil {
			return nil, err
		}
	}

	header, err := p.parseHeader(headerName, component, ctx)
	if err != nil {
		return nil, err
	}

	header.Ref = ref
	p.refs.headers[ref] = header
	return header, nil
}

func (p *parser) resolveExample(ref string, ctx *resolveCtx) (*openapi.Example, error) {
	const prefix = "#/components/examples/"

	if ex, ok := p.refs.examples[ref]; ok {
		return ex, nil
	}

	key, err := ctx.add(ref)
	if err != nil {
		return nil, err
	}
	defer func() {
		ctx.delete(key)
	}()

	var component *ogen.Example
	if key.loc == "" && ctx.lastLoc() == "" {
		name := strings.TrimPrefix(ref, prefix)
		c, found := p.spec.Components.Examples[name]
		if found {
			component = c
		} else {
			if err := resolvePointer(p.spec.Raw, ref, &component); err != nil {
				return nil, err
			}
		}
	} else {
		if err := p.resolve(key, ctx, &component); err != nil {
			return nil, err
		}
	}

	ex, err := p.parseExample(component, ctx)
	if err != nil {
		return nil, err
	}

	if ex != nil {
		ex.Ref = ref
	}
	p.refs.examples[ref] = ex
	return ex, nil
}

func (p *parser) resolveSecurityScheme(ref string, ctx *resolveCtx) (*ogen.SecurityScheme, error) {
	const prefix = "#/components/securitySchemes/"

	if r, ok := p.refs.securitySchemes[ref]; ok {
		return r, nil
	}

	key, err := ctx.add(ref)
	if err != nil {
		return nil, err
	}
	defer func() {
		ctx.delete(key)
	}()

	var component *ogen.SecurityScheme
	if key.loc == "" && ctx.lastLoc() == "" {
		name := strings.TrimPrefix(ref, prefix)
		c, found := p.spec.Components.SecuritySchemes[name]
		if found {
			component = c
		} else {
			if err := resolvePointer(p.spec.Raw, ref, &component); err != nil {
				return nil, err
			}
		}
	} else {
		if err := p.resolve(key, ctx, &component); err != nil {
			return nil, err
		}
	}

	p.refs.securitySchemes[ref] = component
	return component, nil
}
