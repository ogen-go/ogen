package gen

import (
	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/internal/oas"
)

func (g *Generator) generateParameters(opName string, params []*oas.Parameter) ([]*ir.Parameter, error) {
	var result []*ir.Parameter
	for _, p := range params {
		if p.In == oas.LocationCookie {
			err := &ErrNotImplemented{"cookie params"}
			if g.shouldFail(err) {
				return nil, err
			}
			continue
		}

		typ, err := g.generateSchema(pascal(opName, p.Name), p.Schema)
		if err != nil {
			return nil, xerrors.Errorf("%q: %w", p.Name, err)
		}

		visited := map[*ir.Type]struct{}{}
		if err := isParamAllowed(typ, true, visited); err != nil {
			return nil, err
		}

		for t := range visited {
			if t.Is(ir.KindStruct) {
				g.uritypes[t] = struct{}{}
			}
		}

		result = append(result, &ir.Parameter{
			Name: pascal(p.Name),
			Type: typ,
			Spec: p,
		})
	}

	return result, nil
}

func isParamAllowed(t *ir.Type, root bool, visited map[*ir.Type]struct{}) error {
	if _, ok := visited[t]; ok {
		return nil
	}

	visited[t] = struct{}{}
	switch t.Kind {
	case ir.KindPrimitive:
		return nil
	case ir.KindEnum:
		return nil
	case ir.KindArray:
		if !root {
			return xerrors.Errorf("nested arrays not allowed")
		}
		return isParamAllowed(t.Item, false, visited)
	case ir.KindAlias:
		return isParamAllowed(t.AliasTo, root, visited)
	case ir.KindPointer:
		return isParamAllowed(t.PointerTo, root, visited)
	case ir.KindStruct:
		if !root {
			return xerrors.Errorf("nested objects not allowed")
		}
		for _, field := range t.Fields {
			if err := isParamAllowed(field.Type, false, visited); err != nil {
				// TODO: Check field.Spec existence.
				return xerrors.Errorf("field '%s': %w", field.Spec.Name, err)
			}
		}
		return nil
	case ir.KindGeneric:
		return isParamAllowed(t.GenericOf, root, visited)
	case ir.KindSum:
		return &ErrNotImplemented{"sum type parameter"}

		// for i, of := range t.SumOf {
		// 	if err := isParamAllowed(of, root, visited); err != nil {
		// 		// TODO: Check field.Spec existence.
		// 		return xerrors.Errorf("sum[%d]: %w", i, err)
		// 	}
		// }
		// return nil
	default:
		panic("unreachable")
	}
}
