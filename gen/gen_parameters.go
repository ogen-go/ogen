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
			if err := g.fail(&ErrNotImplemented{"cookie params"}); err != nil {
				return nil, errors.Wrap(err, "fail")
			}

			continue
		}

		if err := isSupportedParamStyle(p); err != nil {
			if err := g.fail(err); err != nil {
				return nil, errors.Wrap(err, "fail")
			}

			continue
		}

		t, err := g.generateSchema(pascal(opName, p.Name), p.Schema)
		if err != nil {
			return nil, errors.Wrapf(err, "%q", p.Name)
		}

		visited := map[*ir.Type]struct{}{}
		if err := isParamAllowed(t, true, visited); err != nil {
			return nil, err
		}

		t.AddFeature("uri")
		result = append(result, &ir.Parameter{
			Name: pascal(p.Name),
			Type: t,
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
		return isParamAllowed(t.Item, false, visited)
	case ir.KindAlias:
		return isParamAllowed(t.AliasTo, root, visited)
	case ir.KindPointer:
		return isParamAllowed(t.PointerTo, root, visited)
	case ir.KindStruct:
		if !root {
			return errors.New("nested objects not allowed")
		}
		for _, field := range t.Fields {
			if err := isParamAllowed(field.Type, false, visited); err != nil {
				// TODO: Check field.Spec existence.
				return errors.Wrapf(err, "field %q", field.Spec.Name)
			}
		}
		return nil
	case ir.KindGeneric:
		return isParamAllowed(t.GenericOf, root, visited)
	case ir.KindSum:
		// for i, of := range t.SumOf {
		// 	if err := isParamAllowed(of, false, visited); err != nil {
		// 		// TODO: Check field.Spec existence.
		// 		return errors.Wrapf(err, "sum[%d]", i)
		// 	}
		// }
		// return nil
		return &ErrNotImplemented{"sum type parameter"}
	case ir.KindMap:
		return &ErrNotImplemented{"object with additionalProperties"}
	default:
		panic("unreachable")
	}
}

func isSupportedParamStyle(param *oas.Parameter) error {
	switch param.Style {
	case oas.QueryStyleSpaceDelimited:
		return &ErrNotImplemented{Name: "spaceDelimited parameter style"}

	case oas.QueryStylePipeDelimited:
		if param.Schema.Type == oas.Object {
			return &ErrNotImplemented{Name: "pipeDelimited style for object parameters"}
		}
	}

	return nil
}
