package gen

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/internal/oas"
)

func (g *Generator) generateParameters(opName string, params []*oas.Parameter) ([]*ir.Parameter, error) {
	result := make([]*ir.Parameter, 0, len(params))
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
			return nil, errors.Wrapf(err, "%q", p.Name)
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

	// Params in different locations may have the same names,
	// so we need to resolve name collision in such case.
	for i, p := range result {
		for j, pp := range result {
			if i == j {
				continue
			}

			if p.Name == pp.Name {
				if p.Spec.In == pp.Spec.In {
					panic("unreachable")
				}
				p.Name = string(p.Spec.In) + p.Name
				pp.Name = string(pp.Spec.In) + pp.Name
			}
		}
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
			return errors.New("nested arrays not allowed")
		}
		return isParamAllowed(t.Array.Item, false, visited)
	case ir.KindAlias:
		return isParamAllowed(t.Alias.To, root, visited)
	case ir.KindPointer:
		return isParamAllowed(t.Pointer.To, root, visited)
	case ir.KindStruct:
		if !root {
			return errors.New("nested objects not allowed")
		}
		for _, field := range t.Struct.Fields {
			if err := isParamAllowed(field.Type, false, visited); err != nil {
				// TODO: Check field.Spec existence.
				return errors.Wrapf(err, "field %q", field.Spec.Name)
			}
		}
		return nil
	case ir.KindGeneric:
		return isParamAllowed(t.Generic.Of, root, visited)
	case ir.KindSum:
		return &ErrNotImplemented{"sum type parameter"}

		// for i, of := range t.SumOf {
		// 	if err := isParamAllowed(of, root, visited); err != nil {
		// 		// TODO: Check field.Spec existence.
		// 		return errors.Wrapf(err, "sum[%d]", i)
		// 	}
		// }
		// return nil
	default:
		panic("unreachable")
	}
}
