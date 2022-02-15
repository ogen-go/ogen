package gen

import (
	"strconv"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/internal/oas"
	"github.com/ogen-go/ogen/jsonschema"
)

func (g *Generator) generateParameters(ctx *genctx, opName string, params []*oas.Parameter) ([]*ir.Parameter, error) {
	result := make([]*ir.Parameter, 0, len(params))
	for i, p := range params {
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

		ctx := ctx.appendPath(strconv.Itoa(i), p.Name, "schema")
		t, err := g.generateSchema(ctx, pascal(opName, p.Name), p.Schema)
		if err != nil {
			return nil, errors.Wrapf(err, "%q", p.Name)
		}

		t, err = boxType(ctx, ir.GenericVariant{
			Nullable: p.Schema != nil && p.Schema.Nullable,
			Optional: !p.Required,
		}, t)
		if err != nil {
			return nil, errors.Wrapf(err, "%q", p.Name)
		}

		visited := map[*ir.Type]struct{}{}
		if err := isParamAllowed(t, true, visited); err != nil {
			return nil, err
		}

		t.AddFeature("uri")
		result = append(result, &ir.Parameter{
			Name: pascalNonEmpty(p.Name),
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
				inEqual := p.Spec.In == pp.Spec.In
				specNameEqual := p.Spec.Name == pp.Spec.Name
				switch {
				case inEqual && specNameEqual:
					panic(unreachable(pp.Spec.Name))
				case inEqual:
					p.Name = pascalSpecial(p.Spec.Name)
					pp.Name = pascalSpecial(pp.Spec.Name)
				case specNameEqual:
					p.Name = string(p.Spec.In) + p.Name
					pp.Name = string(pp.Spec.In) + pp.Name
				}
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
	case ir.KindAny:
		return &ErrNotImplemented{"any type parameter"}
	default:
		panic(unreachable(t))
	}
}

func isSupportedParamStyle(param *oas.Parameter) error {
	switch param.Style {
	case oas.QueryStyleSpaceDelimited:
		return &ErrNotImplemented{Name: "spaceDelimited parameter style"}

	case oas.QueryStylePipeDelimited:
		if param.Schema.Type == jsonschema.Object {
			return &ErrNotImplemented{Name: "pipeDelimited style for object parameters"}
		}
	}

	return nil
}
