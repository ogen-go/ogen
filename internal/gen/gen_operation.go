package gen

import (
	"strings"

	"github.com/ogen-go/errors"

	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/internal/oas"
)

func (g *Generator) generateOperation(spec *oas.Operation) (_ *ir.Operation, err error) {
	op := &ir.Operation{
		Name: pascal(spec.Path(), strings.ToLower(spec.HTTPMethod)),
		Spec: spec,
	}
	if spec.OperationID != "" {
		op.Name = pascal(spec.OperationID)
	}

	// Convert []oas.Parameter to []ir.Parameter.
	op.Params, err = g.generateParameters(op.Name, spec.Parameters)
	if err != nil {
		return nil, errors.Wrap(err, "parameters")
	}

	// Convert []oas.PathPart to []ir.PathPart.
	for _, part := range spec.PathParts {
		if part.Raw != "" {
			op.PathParts = append(op.PathParts, &ir.PathPart{Raw: part.Raw})
			continue
		}

		param, found := findParam(op.Params, part.Param.Name)
		if !found {
			panic("unreachable")
		}
		op.PathParts = append(op.PathParts, &ir.PathPart{Param: param})
	}

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

func findParam(params []*ir.Parameter, specName string) (*ir.Parameter, bool) {
	for _, p := range params {
		if p.Spec.Name == specName {
			return p, true
		}
	}
	return nil, false
}
