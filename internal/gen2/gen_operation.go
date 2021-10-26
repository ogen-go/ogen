package gen

import (
	"strings"

	ast "github.com/ogen-go/ogen/internal/ast2"
	"github.com/ogen-go/ogen/internal/ir"
	"golang.org/x/xerrors"
)

func (g *Generator) generateOperation(spec *ast.Operation) (_ *ir.Operation, err error) {
	op := &ir.Operation{
		Name: pascal(spec.Path(), strings.ToLower(spec.HTTPMethod)),
		Spec: spec,
	}
	if spec.OperationID != "" {
		op.Name = pascal(spec.OperationID)
	}

	// Convert []ast.Parameter to []ir.Parameter.
	op.Params, err = g.generateParameters(spec.Parameters)
	if err != nil {
		return nil, xerrors.Errorf("parameters: %w", err)
	}

	// Convert []ast.PathPart to []ir.PathPart.
	for _, part := range spec.PathParts {
		if part.Raw != "" {
			op.PathParts = append(op.PathParts, &ir.PathPart{
				Raw: part.Raw,
			})
			continue
		}

		param, err := g.generateParameter(part.Param)
		if err != nil {
			return nil, xerrors.Errorf("")
		}

		op.PathParts = append(op.PathParts, &ir.PathPart{
			Param: param,
		})
	}

	if spec.RequestBody != nil {
		op.Request, err = g.generateRequest(op.Name+"Request", spec.RequestBody)
		if err != nil {
			return nil, xerrors.Errorf("requestBody: %w", err)
		}
	}

	op.Response, err = g.generateResponses(op.Name+"Response", spec.Responses)
	if err != nil {
		return nil, xerrors.Errorf("responses: %w", err)
	}

	return op, nil
}
