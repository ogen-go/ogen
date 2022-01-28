package gen

import (
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/internal/oas"
)

func (g *Generator) generateOperation(spec *oas.Operation) (_ *ir.Operation, err error) {
	op := &ir.Operation{
		Name:        pascal(spec.Path.String(), strings.ToLower(spec.HTTPMethod)),
		Description: spec.Description,
		Spec:        spec,
	}
	if spec.OperationID != "" {
		op.Name = pascal(spec.OperationID)
	}

	// Convert []oas.Parameter to []*ir.Parameter.
	op.Params, err = g.generateParameters(op.Name, spec.Parameters)
	if err != nil {
		return nil, errors.Wrap(err, "parameters")
	}

	// Convert []oas.PathPart to []*ir.PathPart
	op.PathParts = convertPathParts(op.Spec.Path, op.PathParams())

	if spec.RequestBody != nil {
		op.Request, err = g.generateRequest(op.Name, spec.RequestBody)
		if err != nil {
			return nil, errors.Wrap(err, "requestBody")
		}
	}

	op.Response, err = g.generateResponses(op.Name, spec.Responses)
	if err != nil {
		return nil, errors.Wrap(err, "responses")
	}

	return op, nil
}

func convertPathParts(parts []oas.PathPart, params []*ir.Parameter) []*ir.PathPart {
	find := func(pname string) (*ir.Parameter, bool) {
		for _, p := range params {
			if p.Spec.Name == pname && p.Spec.In == oas.LocationPath {
				return p, true
			}
		}
		return nil, false
	}

	result := make([]*ir.PathPart, 0, len(parts))
	for _, part := range parts {
		if part.Raw != "" {
			result = append(result, &ir.PathPart{Raw: part.Raw})
			continue
		}

		param, found := find(part.Param.Name)
		if !found {
			panic("unreachable")
		}

		result = append(result, &ir.PathPart{Param: param})
	}

	return result
}
