package gen

import (
	"strings"

	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/ast"
)

func (g *Generator) generateParams(methodName string, methodParams []ogen.Parameter) ([]*ast.Parameter, error) {
	var result []*ast.Parameter
	for _, p := range methodParams {
		param, err := g.generateParameter(methodName, p)
		if err != nil {
			return nil, xerrors.Errorf("generate parameter '%s': %w", p.Name, err)
		}

		result = append(result, param)
	}

	// Fix name collisions for parameters in different locations.
	params := make(map[string]*ast.Parameter)
	for _, param := range result {
		p, found := params[param.Name]
		if !found {
			params[param.Name] = param
			continue
		}

		if param.Name == p.Name &&
			param.In == p.In {
			return nil, xerrors.Errorf("parameter name collision: %s", param.Name)
		}

		param.Name = string(param.In) + param.Name
		p.Name = string(p.In) + p.Name
	}

	return result, nil
}

func (g *Generator) generateParameter(name string, param ogen.Parameter) (*ast.Parameter, error) {
	if ref := param.Ref; ref != "" {
		p, err := g.resolveParameter(ref)
		if err != nil {
			return nil, xerrors.Errorf("resolve '%s' reference: %w", ref, err)
		}
		return p, nil
	}

	types := map[string]ast.ParameterLocation{
		"query":  ast.LocationQuery,
		"header": ast.LocationHeader,
		"path":   ast.LocationPath,
		"cookie": ast.LocationCookie,
	}

	locatedIn, exists := types[strings.ToLower(param.In)]
	if !exists {
		return nil, xerrors.Errorf("unsupported parameter type %s", param.In)
	}

	name = pascal(name, param.Name)
	schema, err := g.generateSchema(name, param.Schema)
	if err != nil {
		return nil, xerrors.Errorf("schema: %w", err)
	}

	switch schema.Kind {
	case ast.KindStruct, ast.KindEnum:
		return nil, &ErrNotImplemented{"complex parameter types"}
	case ast.KindArray:
		if !schema.Item.Is(ast.KindPrimitive) {
			return nil, &ErrNotImplemented{"array parameter with complex type"}
		}

		name = pascal(param.Name)
	case ast.KindAlias:
		if !schema.AliasTo.Is(ast.KindPrimitive) {
			return nil, &ErrNotImplemented{"complex parameter types"}
		}

		name = pascal(param.Name)
		schema = schema.AliasTo
	case ast.KindPrimitive:
		name = pascal(param.Name)
	default:
		panic("unreachable")
	}

	style, err := paramStyle(locatedIn, param.Style)
	if err != nil {
		return nil, xerrors.Errorf("style: %w", err)
	}

	return &ast.Parameter{
		Name:       name,
		In:         locatedIn,
		SourceName: param.Name,
		Schema:     schema,
		Style:      style,
		Explode:    paramExplode(locatedIn, param.Explode),
		Required:   param.Required,
	}, nil
}

// paramStyle checks parameter style field.
// https://swagger.io/docs/specification/serialization/
func paramStyle(locatedIn ast.ParameterLocation, style string) (string, error) {
	if style == "" {
		defaultStyles := map[ast.ParameterLocation]string{
			ast.LocationPath:   "simple",
			ast.LocationQuery:  "form",
			ast.LocationHeader: "simple",
			ast.LocationCookie: "form",
		}

		return defaultStyles[locatedIn], nil
	}

	allowedStyles := map[ast.ParameterLocation]map[string]struct{}{
		ast.LocationPath: {
			"simple": struct{}{},
			"label":  struct{}{},
			"matrix": struct{}{},
		},
		ast.LocationQuery: {
			"form": struct{}{},
			// Not supported.
			// "spaceDelimited": struct{}{},
			"pipeDelimited": struct{}{},
			"deepObject":    struct{}{},
		},
		ast.LocationHeader: {
			"simple": struct{}{},
		},
		ast.LocationCookie: {
			"form": struct{}{},
		},
	}

	if _, found := allowedStyles[locatedIn][style]; !found {
		return "", xerrors.Errorf("unexpected style: %s", style)
	}

	return style, nil
}

// paramExplode checks parameter explode field.
// https://swagger.io/docs/specification/serialization/
func paramExplode(locatedIn ast.ParameterLocation, explode *bool) bool {
	if explode != nil {
		return *explode
	}

	// When style is form, the default value is true.
	// For all other styles, the default value is false.
	if locatedIn == ast.LocationQuery || locatedIn == ast.LocationCookie {
		return true
	}

	return false
}
