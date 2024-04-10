package parser

import (
	"fmt"
	"go/token"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonpointer"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/location"
	"github.com/ogen-go/ogen/openapi"
)

type (
	pathItem     = []*openapi.Operation
	unparsedPath struct {
		path string
		loc  location.Locator
		file location.File
	}
)

const (
	xOgenOperationGroup = "x-ogen-operation-group"
)

func (up unparsedPath) String() string {
	return up.path
}

func (p *parser) parsePathItem(
	up unparsedPath,
	item *ogen.PathItem,
	ctx *jsonpointer.ResolveCtx,
) (_ pathItem, rerr error) {
	if item == nil {
		return nil, errors.New("pathItem object is empty or null")
	}
	locator := item.Common.Locator
	defer func() {
		rerr = p.wrapLocation(p.file(ctx), locator, rerr)
	}()

	if ref := item.Ref; ref != "" {
		ops, err := p.resolvePathItem(up, ref, ctx)
		if err != nil {
			return nil, p.wrapRef(p.file(ctx), locator, err)
		}
		return ops, nil
	}

	itemParams, err := p.parseParams(item.Parameters, locator.Field("parameters"), ctx)
	if err != nil {
		return nil, errors.Wrap(err, "parameters")
	}

	// Look for x-ogen-operation-group on the PathItem.
	// Use it as a default value for operations.
	var operationGroup string
	err = p.parseOperationGroup(item.Common, &operationGroup)
	if err != nil {
		return nil, errors.Wrap(err, xOgenOperationGroup)
	}

	var ops []*openapi.Operation
	if err := forEachOps(item, func(method string, op ogen.Operation) error {
		locator := op.Common.Locator
		defer func() {
			rerr = p.wrapLocation(p.file(ctx), locator, rerr)
		}()

		if id := op.OperationID; id != "" {
			ptr := locator.Field("operationId").Pointer(p.file(ctx))
			if existingPtr, ok := p.operationIDs[id]; ok {
				me := new(location.MultiError)
				me.ReportPtr(existingPtr, fmt.Sprintf("duplicate operationId: %q", id))
				me.ReportPtr(ptr, "")
				return me
			}
			p.operationIDs[id] = ptr
		}

		parsedOp, err := p.parseOp(up, method, op, itemParams, ctx, operationGroup)
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
	up unparsedPath,
	httpMethod string,
	spec ogen.Operation,
	itemParams []*openapi.Parameter,
	ctx *jsonpointer.ResolveCtx,
	operationGroup string,
) (_ *openapi.Operation, err error) {
	locator := spec.Common.Locator
	defer func() {
		err = p.wrapLocation(p.file(ctx), locator, err)
	}()

	op := &openapi.Operation{
		OperationID:         spec.OperationID,
		Summary:             spec.Summary,
		Description:         spec.Description,
		Deprecated:          spec.Deprecated,
		HTTPMethod:          httpMethod,
		Pointer:             locator.Pointer(p.file(ctx)),
		XOgenOperationGroup: operationGroup,
	}

	err = p.parseOperationGroup(spec.Common, &op.XOgenOperationGroup)
	if err != nil {
		return nil, errors.Wrap(err, xOgenOperationGroup)
	}

	opParams, err := p.parseParams(spec.Parameters, locator.Field("parameters"), ctx)
	if err != nil {
		return nil, errors.Wrap(err, "parameters")
	}

	// Merge operation parameters with pathItem parameters.
	op.Parameters = mergeParams(opParams, itemParams)

	op.Path, err = parsePath(up.path, op.Parameters)
	if err != nil {
		// Special case: point to the operation "parameters" what caused the error.
		// It is helpful, since one Path Item may contain multiple operations.
		if pe, ok := errors.Into[*pathParameterNotSpecifiedError](err); ok {
			return nil, p.wrapField("parameters", p.file(ctx), locator, pe)
		}

		err := errors.Wrapf(err, "parse path %q", up)
		return nil, p.wrapLocation(up.file, up.loc, err)
	}

	if spec.RequestBody != nil {
		op.RequestBody, err = p.parseRequestBody(spec.RequestBody, ctx)
		if err != nil {
			return nil, errors.Wrap(err, "requestBody")
		}
	}

	{
		locator := locator.Field("responses")
		op.Responses, err = p.parseResponses(spec.Responses, locator, ctx)
		if err != nil {
			err := errors.Wrap(err, "responses")
			return nil, p.wrapLocation(p.file(ctx), locator, err)
		}
	}

	parseSecurity := func(spec ogen.SecurityRequirements, locator location.Locator) (err error) {
		op.Security, err = p.parseSecurityRequirements(spec, locator, ctx)
		if err != nil {
			err := errors.Wrap(err, "security")
			return p.wrapLocation(p.file(ctx), locator, err)
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
		securityParent = locator
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

func (p *parser) parseOperationGroup(common jsonschema.OpenAPICommon, operationGroup *string) error {
	if ex, ok := common.Extensions[xOgenOperationGroup]; ok {
		if err := ex.Decode(operationGroup); err != nil {
			return errors.Wrap(err, "unmarshal value")
		}

		if !token.IsIdentifier(*operationGroup) {
			return errors.Errorf("%q is not a valid identifier", *operationGroup)
		}
	}

	return nil
}
