package gen

import (
	"strings"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

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

	if spec.XOgenOperationGroup != "" {
		op.OperationGroup, err = pascalNonEmpty(spec.XOgenOperationGroup)
		if err != nil {
			return nil, errors.Wrap(err, "operation group")
		}
	}

	vetPathParametersUsed(g.log, op.Spec.Path, spec.Parameters)
	// Convert []openapi.Parameter to []*ir.Parameter.
	op.Params, err = g.generateParameters(ctx, op.Name, spec.Parameters)
	if err != nil {
		return nil, errors.Wrap(err, "parameters")
	}

	// Convert []openapi.PathPart to []*ir.PathPart
	op.PathParts = convertPathParts(op.Spec.Path, op.Params)

	if spec.RequestBody != nil {
		op.Request, err = g.generateRequest(ctx, op.Name, spec.RequestBody)
		if err != nil {
			return nil, errors.Wrap(err, "requestBody")
		}
	}

	op.Responses, err = g.generateResponses(ctx, op.Name, spec.Responses)
	if err != nil {
		return nil, errors.Wrap(err, "responses")
	}

	op.Security, err = g.generateSecurities(ctx, opName, spec.Security)
	if err != nil {
		return nil, errors.Wrap(err, "security")
	}

	return op, nil
}

func vetPathParametersUsed(log *zap.Logger, parts openapi.Path, params []*openapi.Parameter) {
	used := map[string]struct{}{}
	for _, p := range parts {
		if !p.IsParam() || !p.Param.In.Path() {
			continue
		}
		used[p.Param.Name] = struct{}{}
	}

	for _, p := range params {
		if !p.In.Path() {
			continue
		}
		if _, ok := used[p.Name]; !ok {
			log.Warn("Path parameter is not used",
				zap.String("name", p.Name),
				zapPosition(p),
			)
		}
	}
}

func convertPathParts(parts openapi.Path, params []*ir.Parameter) []*ir.PathPart {
	find := func(pname string) (*ir.Parameter, bool) {
		return xslices.FindFunc(params, func(p *ir.Parameter) bool {
			return p.Spec.Name == pname && p.Spec.In.Path()
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
