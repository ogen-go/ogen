package gen

import (
	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/internal/oas"
)

func (g *Generator) generateParameters(opName string, params []*oas.Parameter) ([]*ir.Parameter, error) {
	var result []*ir.Parameter
	for _, p := range params {
		typ, err := g.generateSchema(pascal(opName, p.Name), p.Schema)
		if err != nil {
			return nil, xerrors.Errorf("'%s': %w", p.Name, err)
		}

		if !isUnderlyingPrimitive(typ) {
			return nil, &ErrNotImplemented{"complex parameter types"}
		}

		result = append(result, &ir.Parameter{
			Name: pascal(p.Name),
			Type: typ,
			Spec: p,
		})
	}

	return result, nil
}

func isUnderlyingPrimitive(typ *ir.Type) bool {
	switch typ.Kind {
	case ir.KindPrimitive, ir.KindEnum:
		return true
	case ir.KindArray:
		return isUnderlyingPrimitive(typ.Item)
	case ir.KindAlias:
		return isUnderlyingPrimitive(typ.AliasTo)
	case ir.KindPointer:
		return isUnderlyingPrimitive(typ.PointerTo)
	default:
		return false
	}
}
