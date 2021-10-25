package gen

import (
	"strings"

	ast "github.com/ogen-go/ogen/internal/ast2"
	"github.com/ogen-go/ogen/internal/ir"
	"golang.org/x/xerrors"
)

func (g *Generator) generateMethod(spec *ast.Method) (_ *ir.Method, err error) {
	m := &ir.Method{
		Name: pascal(spec.Path(), strings.ToLower(spec.HTTPMethod)),
		Spec: spec,
	}
	if spec.OperationID != "" {
		m.Name = pascal(spec.OperationID)
	}

	m.Params, err = g.generateParameters(m.Name+"Param", spec.Parameters)
	if err != nil {
		return nil, xerrors.Errorf("parameters: %w", err)
	}

	if spec.RequestBody != nil {
		m.Request, err = g.generateRequest(m.Name+"Request", spec.RequestBody)
		if err != nil {
			return nil, xerrors.Errorf("requestBody: %w", err)
		}
	}

	m.Response, err = g.generateResponses(m.Name+"Response", spec.Responses)
	if err != nil {
		return nil, xerrors.Errorf("responses: %w", err)
	}

	return m, nil
}
