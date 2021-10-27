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
			return nil, xerrors.Errorf("%q: %w", p.Name, err)
		}

		typ, ok := unwrapPrimitive(typ)
		if !ok {
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

func unwrapPrimitive(typ *ir.Type) (*ir.Type, bool) {
	switch typ.Kind {
	case ir.KindPrimitive:
		return typ, true
	case ir.KindEnum:
		return &ir.Type{
			Kind:      ir.KindPrimitive,
			Primitive: typ.Primitive,
		}, true
	case ir.KindArray:
		item, ok := unwrapPrimitive(typ.Item)
		if !ok {
			return nil, false
		}

		typ.Item = item
		return typ, true
	case ir.KindAlias:
		return unwrapPrimitive(typ.AliasTo)
	case ir.KindPointer:
		return unwrapPrimitive(typ.PointerTo)
	default:
		return nil, false
	}
}
