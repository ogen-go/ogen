package gen

import (
	"strconv"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
)

func (g *Generator) generateParameters(ctx *genctx, opName string, params []*openapi.Parameter) (_ []*ir.Parameter, err error) {
	result := make([]*ir.Parameter, 0, len(params))
	for i, p := range params {
		if p.Content != nil {
			if err := g.fail(&ErrNotImplemented{"parameter content field"}); err != nil {
				return nil, errors.Wrap(err, "fail")
			}

			continue
		}

		if p.In == openapi.LocationCookie {
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

		paramTypeName, err := pascal(opName, p.Name)
		if err != nil {
			return nil, errors.Wrapf(err, "parameter type name: %q", p.Name)
		}
		t, err := g.generateSchema(ctx, paramTypeName, p.Schema)
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
		paramName, err := pascalNonEmpty(p.Name)
		if err != nil {
			return nil, errors.Wrapf(err, "parameter name: %q", p.Name)
		}
		result = append(result, &ir.Parameter{
			Name: paramName,
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
					p.Name, err = pascalSpecial(p.Spec.Name)
					if err != nil {
						return nil, errors.Wrap(err, "parameter name")
					}

					pp.Name, err = pascalSpecial(pp.Spec.Name)
					if err != nil {
						return nil, errors.Wrap(err, "parameter name")
					}
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
				return errors.Wrapf(err, "field %q", field.Name)
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

func isSupportedParamStyle(param *openapi.Parameter) error {
	switch param.Style {
	case openapi.QueryStyleSpaceDelimited:
		return &ErrNotImplemented{Name: "spaceDelimited parameter style"}

	case openapi.QueryStylePipeDelimited:
		if param.Schema.Type == jsonschema.Object {
			return &ErrNotImplemented{Name: "pipeDelimited style for object parameters"}
		}
	}

	return nil
}
