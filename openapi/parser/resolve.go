package parser

import (
	"context"
	"strings"

	"github.com/go-faster/errors"
	yaml "github.com/go-faster/yamlx"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/jsonpointer"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) getSchema(ctx *jsonpointer.ResolveCtx) (*yaml.Node, error) {
	loc := ctx.LastLoc()

	r, ok := p.schemas[loc]
	if ok {
		return r, nil
	}

	var node yaml.Node
	if err := func() error {
		raw, err := p.external.Get(context.TODO(), loc)
		if err != nil {
			return errors.Wrap(err, "get")
		}

		if err := yaml.Unmarshal(raw, &node); err != nil {
			return errors.Wrap(err, "unmarshal")
		}

		return nil
	}(); err != nil {
		return nil, errors.Wrapf(err, "external %q", loc)
	}
	r = &node
	p.schemas[loc] = r
	return r, nil
}

func resolvePointer(root *yaml.Node, ptr string, to any) error {
	n, err := jsonpointer.Resolve(ptr, root)
	if err != nil {
		return err
	}
	return n.Decode(to)
}

func (p *parser) resolve(key jsonpointer.RefKey, ctx *jsonpointer.ResolveCtx, to any) error {
	schema, err := p.getSchema(ctx)
	if err != nil {
		return err
	}
	return resolvePointer(schema, key.Ref, to)
}

func (p *parser) resolveRequestBody(ref string, ctx *jsonpointer.ResolveCtx) (*openapi.RequestBody, error) {
	const prefix = "#/components/requestBodies/"

	if r, ok := p.refs.requestBodies[ref]; ok {
		return r, nil
	}

	key, err := ctx.Add(ref)
	if err != nil {
		return nil, err
	}
	defer func() {
		ctx.Delete(key)
	}()

	var component *ogen.RequestBody
	if key.Loc == "" && ctx.LastLoc() == "" {
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

func (p *parser) resolveResponse(ref string, ctx *jsonpointer.ResolveCtx) (*openapi.Response, error) {
	const prefix = "#/components/responses/"

	if r, ok := p.refs.responses[ref]; ok {
		return r, nil
	}

	key, err := ctx.Add(ref)
	if err != nil {
		return nil, err
	}
	defer func() {
		ctx.Delete(key)
	}()

	var component *ogen.Response
	if key.Loc == "" && ctx.LastLoc() == "" {
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

func (p *parser) resolveParameter(ref string, ctx *jsonpointer.ResolveCtx) (*openapi.Parameter, error) {
	const prefix = "#/components/parameters/"

	if param, ok := p.refs.parameters[ref]; ok {
		return param, nil
	}

	key, err := ctx.Add(ref)
	if err != nil {
		return nil, err
	}
	defer func() {
		ctx.Delete(key)
	}()

	var component *ogen.Parameter
	if key.Loc == "" && ctx.LastLoc() == "" {
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

func (p *parser) resolveHeader(headerName, ref string, ctx *jsonpointer.ResolveCtx) (*openapi.Header, error) {
	const prefix = "#/components/headers/"

	if header, ok := p.refs.headers[ref]; ok {
		return header, nil
	}

	key, err := ctx.Add(ref)
	if err != nil {
		return nil, err
	}
	defer func() {
		ctx.Delete(key)
	}()

	var component *ogen.Header
	if key.Loc == "" && ctx.LastLoc() == "" {
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

func (p *parser) resolveExample(ref string, ctx *jsonpointer.ResolveCtx) (*openapi.Example, error) {
	const prefix = "#/components/examples/"

	if ex, ok := p.refs.examples[ref]; ok {
		return ex, nil
	}

	key, err := ctx.Add(ref)
	if err != nil {
		return nil, err
	}
	defer func() {
		ctx.Delete(key)
	}()

	var component *ogen.Example
	if key.Loc == "" && ctx.LastLoc() == "" {
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

func (p *parser) resolveSecurityScheme(ref string, ctx *jsonpointer.ResolveCtx) (*ogen.SecurityScheme, error) {
	const prefix = "#/components/securitySchemes/"

	if r, ok := p.refs.securitySchemes[ref]; ok {
		return r, nil
	}

	key, err := ctx.Add(ref)
	if err != nil {
		return nil, err
	}
	defer func() {
		ctx.Delete(key)
	}()

	var component *ogen.SecurityScheme
	if key.Loc == "" && ctx.LastLoc() == "" {
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

	ss, err := p.parseSecurityScheme(component, ctx)
	if err != nil {
		return nil, err
	}

	p.refs.securitySchemes[ref] = ss
	return ss, nil
}

func (p *parser) resolvePathItem(
	itemPath, ref string,
	operationIDs map[string]struct{},
	ctx *jsonpointer.ResolveCtx,
) (pathItem, error) {
	const prefix = "#/components/pathItems/"

	if r, ok := p.refs.pathItems[ref]; ok {
		return r, nil
	}

	key, err := ctx.Add(ref)
	if err != nil {
		return nil, err
	}
	defer func() {
		ctx.Delete(key)
	}()

	var component *ogen.PathItem
	if key.Loc == "" && ctx.LastLoc() == "" {
		name := strings.TrimPrefix(ref, prefix)
		c, found := p.spec.Components.PathItems[name]
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

	r, err := p.parsePathItem(itemPath, component, operationIDs, ctx)
	if err != nil {
		return nil, err
	}

	p.refs.pathItems[ref] = r
	return r, nil
}
