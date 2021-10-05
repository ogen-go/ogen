package gen

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/ast"
)

func (g *Generator) generateParams(methodPath string, methodParams []ogen.Parameter) ([]*ast.Parameter, error) {
	var result []*ast.Parameter
	for _, p := range methodParams {
		if p.Ref != "" {
			componentParam, found := g.componentsParameter(p.Ref)
			if !found {
				return nil, xerrors.Errorf("parameter by reference '%s' not found", p.Ref)
			}

			p = componentParam
		}

		param, err := g.parseParameter(p, methodPath)
		if xerrors.Is(err, errSkipSchema) {
			continue
		}
		if err != nil {
			return nil, xerrors.Errorf("parse parameter '%s': %w", p.Name, err)
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

func (g *Generator) parseParameter(param ogen.Parameter, path string) (*ast.Parameter, error) {
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

	if locatedIn == ast.LocationPath {
		if !param.Required {
			return nil, xerrors.Errorf("parameters located in 'path' must be required")
		}

		exists, err := regexp.MatchString(fmt.Sprintf("{%s}", param.Name), path)
		if err != nil {
			return nil, xerrors.Errorf("match path param '%s': %w", param.Name, err)
		}

		if !exists {
			return nil, xerrors.Errorf("param '%s' not found in path '%s'", param.Name, path)
		}
	}

	name := pascal(param.Name)
	schema, err := g.generateSchema(name, param.Schema)
	if err != nil {
		return nil, xerrors.Errorf("schema: %w", err)
	}

	switch schema.Kind {
	case ast.KindStruct:
		return nil, xerrors.Errorf("object type not supported")
	case ast.KindArray:
		if schema.Item.Kind != ast.KindPrimitive {
			return nil, xerrors.Errorf("only arrays with primitive types supported")
		}
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
			"form":           struct{}{},
			"spaceDelimited": struct{}{},
			"pipeDelimited":  struct{}{},
			"deepObject":     struct{}{},
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
