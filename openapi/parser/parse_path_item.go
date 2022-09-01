package parser

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/jsonpointer"
	"github.com/ogen-go/ogen/internal/location"
	"github.com/ogen-go/ogen/openapi"
)

type pathItem = []*openapi.Operation

func (p *parser) parsePathItem(
	path string,
	item *ogen.PathItem,
	operationIDs map[string]struct{},
	ctx *jsonpointer.ResolveCtx,
) (_ pathItem, rerr error) {
	if item == nil {
		return nil, errors.New("pathItem object is empty or null")
	}
	defer func() {
		rerr = p.wrapLocation(ctx.LastLoc(), item.Locator, rerr)
	}()

	if ref := item.Ref; ref != "" {
		ops, err := p.resolvePathItem(path, ref, operationIDs, ctx)
		if err != nil {
			return nil, p.wrapRef(ctx.LastLoc(), item.Locator, err)
		}
		return ops, nil
	}

	itemParams, err := p.parseParams(item.Parameters, item.Locator.Field("parameters"), ctx)
	if err != nil {
		return nil, errors.Wrap(err, "parameters")
	}

	var ops []*openapi.Operation
	if err := forEachOps(item, func(method string, op ogen.Operation) error {
		if id := op.OperationID; id != "" {
			if _, ok := operationIDs[id]; ok {
				return errors.Errorf("duplicate operationId: %q", id)
			}
			operationIDs[id] = struct{}{}
		}

		parsedOp, err := p.parseOp(path, method, op, itemParams, ctx)
		if err != nil {
			if op.OperationID != "" {
				return errors.Wrapf(err, "operation %q", op.OperationID)
			}
			return err
		}

		ops = append(ops, parsedOp)
		return nil
	}); err != nil {
		return nil, err
	}

	return ops, nil
}

func (p *parser) parseOp(
	path, httpMethod string,
	spec ogen.Operation,
	itemParams []*openapi.Parameter,
	ctx *jsonpointer.ResolveCtx,
) (_ *openapi.Operation, err error) {
	defer func() {
		err = p.wrapLocation(ctx.LastLoc(), spec.Locator, err)
	}()

	op := &openapi.Operation{
		OperationID: spec.OperationID,
		Summary:     spec.Summary,
		Description: spec.Description,
		Deprecated:  spec.Deprecated,
		HTTPMethod:  httpMethod,
		Locator:     spec.Locator,
	}

	opParams, err := p.parseParams(spec.Parameters, spec.Locator.Field("parameters"), ctx)
	if err != nil {
		return nil, errors.Wrap(err, "parameters")
	}

	// Merge operation parameters with pathItem parameters.
	op.Parameters = mergeParams(opParams, itemParams)

	op.Path, err = parsePath(path, op.Parameters)
	if err != nil {
		return nil, errors.Wrapf(err, "parse path %q", path)
	}

	if spec.RequestBody != nil {
		op.RequestBody, err = p.parseRequestBody(spec.RequestBody, ctx)
		if err != nil {
			return nil, errors.Wrap(err, "requestBody")
		}
	}

	{
		locator := spec.Locator.Field("responses")
		op.Responses, err = p.parseResponses(spec.Responses, locator, ctx)
		if err != nil {
			err := errors.Wrap(err, "responses")
			return nil, p.wrapLocation(ctx.LastLoc(), locator, err)
		}
	}

	parseSecurity := func(spec ogen.SecurityRequirements, locator location.Locator) (err error) {
		op.Security, err = p.parseSecurityRequirements(spec, locator, ctx)
		if err != nil {
			err := errors.Wrap(err, "security")
			return p.wrapLocation(ctx.LastLoc(), locator, err)
		}
		return nil
	}

	var (
		security       = p.spec.Security
		securityParent = p.rootLoc
	)
	if spec.Security != nil {
		// Use operation level security.
		security = spec.Security
		securityParent = spec.Locator
	}
	if err := parseSecurity(security, securityParent.Field("security")); err != nil {
		return nil, err
	}

	return op, nil
}

func forEachOps(item *ogen.PathItem, f func(method string, op ogen.Operation) error) error {
	var err error
	handle := func(method string, op *ogen.Operation) {
		if err != nil || op == nil {
			return
		}

		err = f(method, *op)
		if err != nil {
			err = errors.Wrap(err, method)
		}
	}

	handle("get", item.Get)
	handle("put", item.Put)
	handle("post", item.Post)
	handle("delete", item.Delete)
	handle("options", item.Options)
	handle("head", item.Head)
	handle("patch", item.Patch)
	handle("trace", item.Trace)
	return err
}
