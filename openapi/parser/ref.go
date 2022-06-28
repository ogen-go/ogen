package parser

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonpointer"
	"github.com/ogen-go/ogen/openapi"
)

type refKey struct {
	loc string
	ref string
}

func (r *refKey) fromURL(u *url.URL) {
	{
		// Make copy.
		u2 := *u
		u2.Fragment = ""
		r.loc = u2.String()
	}
	r.ref = "#" + u.Fragment
}

type resolveCtx struct {
	// Location stack. Used for context-depending resolving.
	//
	// For resolve trace like
	//
	// 	"#/components/schemas/Schema" ->
	// 	"https://example.com/schema#Schema" ->
	//	"#/definitions/SchemaProperty"
	//
	// "#/definitions/SchemaProperty" should be resolved against "https://example.com/schema".
	locstack []string
	// Store references to detect infinite recursive references.
	refs       map[refKey]struct{}
	depthLimit int
}

func newResolveCtx(depthLimit int) *resolveCtx {
	return &resolveCtx{
		locstack:   nil,
		refs:       map[refKey]struct{}{},
		depthLimit: depthLimit,
	}
}

func (r *resolveCtx) add(ref string) (key refKey, _ error) {
	u, err := url.Parse(ref)
	if err != nil {
		return refKey{}, err
	}
	key.fromURL(u)

	if r.depthLimit <= 0 {
		return refKey{}, errors.New("depth limit exceeded")
	}
	if _, ok := r.refs[key]; ok {
		return key, errors.New("infinite recursion")
	}
	r.refs[key] = struct{}{}
	r.depthLimit--

	if key.loc != "" {
		r.locstack = append(r.locstack, key.loc)
	}
	return key, nil
}

func (r *resolveCtx) delete(key refKey) {
	r.depthLimit++
	delete(r.refs, key)
	if key.loc != "" && len(r.locstack) > 0 {
		r.locstack = r.locstack[:len(r.locstack)-1]
	}
}

func (r *resolveCtx) lastLoc() string {
	s := r.locstack
	if len(s) == 0 {
		return ""
	}
	return s[len(s)-1]
}

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

func (p *parser) resolve(key refKey, ctx *resolveCtx, to interface{}) error {
	schema, err := p.getSchema(ctx)
	if err != nil {
		return err
	}

	data, err := jsonpointer.Resolve(key.ref, schema)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, to)
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
		if !found {
			return nil, errors.New("component by reference not found")
		}
		component = c
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
		if !found {
			return nil, errors.New("component by reference not found")
		}
		component = c
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
		if !found {
			return nil, errors.New("component by reference not found")
		}
		component = c
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
		if !found {
			return nil, errors.New("component by reference not found")
		}
		component = c
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
		if !found {
			return nil, errors.New("component by reference not found")
		}
		component = c
	} else {
		if err := p.resolve(key, ctx, &component); err != nil {
			return nil, err
		}
	}

	ex, err := p.parseExample(component, ctx)
	if err != nil {
		return nil, err
	}

	ex.Ref = ref
	p.refs.examples[ref] = ex
	return ex, nil
}

func (p *parser) resolveSecuritySchema(ref string, ctx *resolveCtx) (*ogen.SecuritySchema, error) {
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

	var component *ogen.SecuritySchema
	if key.loc == "" && ctx.lastLoc() == "" {
		name := strings.TrimPrefix(ref, prefix)
		c, found := p.spec.Components.SecuritySchemes[name]
		if !found {
			return nil, errors.New("component by reference not found")
		}
		component = c
	} else {
		if err := p.resolve(key, ctx, &component); err != nil {
			return nil, err
		}
	}

	p.refs.securitySchemes[ref] = component
	return component, nil
}
