package gen

import (
	"fmt"
	"net/http"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/internal/naming"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
)

func vetHeaderParameterName(log *zap.Logger, name string, loc position, ignore ...string) (skip bool) {
	canonical := http.CanonicalHeaderKey(name)
	if canonical != name {
		log.Warn(
			"Header name is not canonical, canonical name will be used",
			zapPosition(loc),
			zap.String("original_name", name),
			zap.String("canonical_name", canonical),
		)
	}
	for _, ign := range ignore {
		if ign == canonical {
			skip = true
			log.Warn(
				fmt.Sprintf("%s is described separately and will be ignored in this section", ign),
				zapPosition(loc),
			)
			break
		}
	}
	return skip
}

func (g *Generator) generateParameters(ctx *genctx, opName string, params []*openapi.Parameter) (_ []*ir.Parameter, err error) {
	result := make([]*ir.Parameter, 0, len(params))
	for _, p := range params {
		if p.In.Header() {
			if vetHeaderParameterName(g.log, p.Name, p, "Content-Type", "Authorization") {
				continue
			}
		}

		param, err := g.generateParameter(ctx, opName, p)
		if err != nil {
			if err := g.trySkip(err, "Skipping parameter", p); err != nil {
				return nil, err
			}
			// Path parameters are required.
			if p.In.Path() {
				return nil, err
			}
			continue
		}

		result = append(result, param)
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

					if p.Name == pp.Name {
						return nil, &ErrNotImplemented{"too similar parameter name"}
					}
				default:
					p.Name = naming.Capitalize(p.Spec.In.String()) + p.Name
					pp.Name = naming.Capitalize(pp.Spec.In.String()) + pp.Name
				}
			}
		}
	}

	return result, nil
}

func (g *Generator) generateParameter(ctx *genctx, opName string, p *openapi.Parameter) (ret *ir.Parameter, rerr error) {
	if err := isSupportedParamStyle(p); err != nil {
		return nil, err
	}

	var paramTypeName string
	if ref := p.Ref; !ref.IsZero() {
		if p, ok := ctx.lookupParameter(ref); ok {
			return p, nil
		}

		n, err := pascal(cleanRef(ref))
		if err != nil {
			return nil, errors.Wrapf(err, "parameter type name: %q", ref)
		}
		paramTypeName = n

		defer func() {
			if rerr != nil {
				return
			}

			if err := ctx.saveParameter(ref, ret); err != nil {
				rerr = err
				ret = nil
			}
		}()
	}

	if paramTypeName == "" {
		var err error
		paramTypeName, err = pascal(opName, p.Name)
		if err != nil {
			return nil, errors.Wrapf(err, "parameter type name: %q", p.Name)
		}
	}

	generate := func(ctx *genctx, sch *jsonschema.Schema) (*ir.Type, error) {
		return g.generateSchema(ctx, paramTypeName, sch, !p.Required, nil)
	}
	t, err := func() (*ir.Type, error) {
		if content := p.Content; content != nil {
			if val := content.Name; val != "application/json" {
				return nil, errors.Wrapf(
					&ErrNotImplemented{"parameter content encoding"},
					"%q", val,
				)
			}

			t, err := generate(ctx, content.Media.Schema)
			if err != nil {
				return nil, err
			}

			t.AddFeature("json")
			return t, nil
		}

		t, err := generate(ctx, p.Schema)
		if err != nil {
			return nil, err
		}

		visited := map[*ir.Type]struct{}{}
		if err := isParamAllowed(t, true, visited); err != nil {
			return nil, err
		}

		t.AddFeature("uri")
		return t, nil
	}()
	if err != nil {
		return nil, errors.Wrapf(err, "%q", p.Name)
	}

	paramName, err := pascalNonEmpty(p.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "parameter name: %q", p.Name)
	}

	return &ir.Parameter{
		Name: paramName,
		Type: t,
		Spec: p,
	}, nil
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
		if s := param.Schema; s != nil && s.Type == jsonschema.Object {
			return &ErrNotImplemented{Name: "pipeDelimited style for object parameters"}
		}
	}
	return nil
}
