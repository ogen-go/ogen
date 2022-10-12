package gen

import (
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/internal/xslices"
	"github.com/ogen-go/ogen/openapi"
)

func (g *Generator) generateOperation(ctx *genctx, webhookName string, spec *openapi.Operation) (_ *ir.Operation, err error) {
	var opName string
	switch {
	case spec.OperationID != "":
		opName, err = pascalNonEmpty(spec.OperationID)
	case webhookName != "":
		opName, err = pascalNonEmpty(webhookName, strings.ToLower(spec.HTTPMethod))
	default:
		opName, err = pascal(spec.Path.String(), strings.ToLower(spec.HTTPMethod))
	}
	if err != nil {
		return nil, errors.Wrap(err, "operation name")
	}

	op := &ir.Operation{
		Name:        opName,
		Summary:     spec.Summary,
		Description: spec.Description,
		Deprecated:  spec.Deprecated,
		Spec:        spec,
	}

	// Convert []openapi.Parameter to []*ir.Parameter.
	op.Params, err = g.generateParameters(ctx.appendPath("parameters"), op.Name, spec.Parameters)
	if err != nil {
		return nil, errors.Wrap(err, "parameters")
	}

	// Convert []openapi.PathPart to []*ir.PathPart
	op.PathParts = convertPathParts(op.Spec.Path, op.PathParams())

	if spec.RequestBody != nil {
		op.Request, err = g.generateRequest(ctx.appendPath("requestBody"), op.Name, spec.RequestBody)
		if err != nil {
			return nil, errors.Wrap(err, "requestBody")
		}
	}

	op.Responses, err = g.generateResponses(ctx.appendPath("responses"), op.Name, spec.Responses)
	if err != nil {
		return nil, errors.Wrap(err, "responses")
	}

	op.Security, err = g.generateSecurities(ctx.appendPath("security"), spec.Security)
	if err != nil {
		return nil, errors.Wrap(err, "security")
	}

	return op, nil
}

func convertPathParts(parts openapi.Path, params []*ir.Parameter) []*ir.PathPart {
	find := func(pname string) (*ir.Parameter, bool) {
		return xslices.FindFunc(params, func(p *ir.Parameter) bool {
			return p.Spec.Name == pname && p.Spec.In == openapi.LocationPath
		})
	}

	result := make([]*ir.PathPart, 0, len(parts))
	for _, part := range parts {
		if !part.IsParam() {
			result = append(result, &ir.PathPart{Raw: part.Raw})
			continue
		}

		param, found := find(part.Param.Name)
		if !found {
			panic(unreachable(part.Param.Name))
		}

		result = append(result, &ir.PathPart{Param: param})
	}

	return result
}
