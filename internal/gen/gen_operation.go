package gen

import (
	"strings"

	"golang.org/x/xerrors"

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
	op.Params, err = g.generateParameters(spec.Parameters)
	if err != nil {
		return nil, xerrors.Errorf("parameters: %w", err)
	}

	// Convert []oas.PathPart to []ir.PathPart.
	for _, part := range spec.PathParts {
		if part.Raw != "" {
			op.PathParts = append(op.PathParts, &ir.PathPart{
				Raw: part.Raw,
			})
			continue
		}

		param, err := g.generateParameter(part.Param)
		if err != nil {
			return nil, xerrors.Errorf("param: %w", err)
		}

		op.PathParts = append(op.PathParts, &ir.PathPart{
			Param: param,
		})
	}

	if spec.RequestBody != nil {
		op.Request, err = g.generateRequest(op.Name+"Req", spec.RequestBody)
		if err != nil {
			return nil, xerrors.Errorf("requestBody: %w", err)
		}
	}

	op.Response, err = g.generateResponses(op.Name+"Res", spec.Responses)
	if err != nil {
		return nil, xerrors.Errorf("responses: %w", err)
	}

	return op, nil
}
