package gen

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ogen-go/ogen"
)

func (g *Generator) generateParams(methodPath string, methodParams []ogen.Parameter) ([]*Parameter, error) {
	var result []*Parameter
	for _, p := range methodParams {
		if p.Ref != "" {
			componentParam, found := g.componentsParameter(p.Ref)
			if !found {
				return nil, fmt.Errorf("parameter by reference '%s' not found", p.Ref)
			}

			p = componentParam
		}

		param, err := g.parseParameter(p, methodPath)
		if err != nil {
			return nil, fmt.Errorf("parse parameter '%s': %w", p.Name, err)
		}

		result = append(result, param)
	}

	// Fix name collisions for parameters in different locations.
	params := make(map[string]*Parameter)
	for _, param := range result {
		p, found := params[param.Name]
		if !found {
			params[param.Name] = param
			continue
		}

		if param.Name == p.Name &&
			param.In == p.In {
			return nil, fmt.Errorf("parameter name collision: %s", param.Name)
		}

		param.Name = string(param.In) + param.Name
		p.Name = string(p.In) + p.Name
	}

	return result, nil
}

func (g *Generator) parseParameter(param ogen.Parameter, path string) (*Parameter, error) {
	types := map[string]ParameterLocation{
		"query":  LocationQuery,
		"header": LocationHeader,
		"path":   LocationPath,
		"cookie": LocationCookie,
	}

	locatedIn, exists := types[strings.ToLower(param.In)]
	if !exists {
		return nil, fmt.Errorf("unsupported parameter type %s", param.In)
	}

	if locatedIn == LocationPath {
		if !param.Required {
			return nil, fmt.Errorf("parameters located in 'path' must be required")
		}

		exists, err := regexp.MatchString(fmt.Sprintf("{%s}", param.Name), path)
		if err != nil {
			return nil, fmt.Errorf("match path param '%s': %w", param.Name, err)
		}

		if !exists {
			return nil, fmt.Errorf("param '%s' not found in path '%s'", param.Name, path)
		}
	}

	name := pascal(param.Name)
	schema, err := g.generateSchema(name, param.Schema)
	if err != nil {
		return nil, fmt.Errorf("schema: %w", err)
	}

	switch schema.Kind {
	case KindStruct:
		return nil, fmt.Errorf("object type not supported")
	case KindArray:
		if schema.Item.Kind != KindPrimitive {
			return nil, fmt.Errorf("only arrays with primitive types supported")
		}
	}

	style, err := paramStyle(locatedIn, param.Style)
	if err != nil {
		return nil, fmt.Errorf("style: %w", err)
	}

	return &Parameter{
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
func paramStyle(locatedIn ParameterLocation, style string) (string, error) {
	if style == "" {
		defaultStyles := map[ParameterLocation]string{
			LocationPath:   "simple",
			LocationQuery:  "form",
			LocationHeader: "simple",
			LocationCookie: "form",
		}

		return defaultStyles[locatedIn], nil
	}

	allowedStyles := map[ParameterLocation]map[string]struct{}{
		LocationPath: {
			"simple": struct{}{},
			"label":  struct{}{},
			"matrix": struct{}{},
		},
		LocationQuery: {
			"form":           struct{}{},
			"spaceDelimited": struct{}{},
			"pipeDelimited":  struct{}{},
			"deepObject":     struct{}{},
		},
		LocationHeader: {
			"simple": struct{}{},
		},
		LocationCookie: {
			"form": struct{}{},
		},
	}

	if _, found := allowedStyles[locatedIn][style]; !found {
		return "", fmt.Errorf("unexpected style: %s", style)
	}

	return style, nil
}

// paramExplode checks parameter explode field.
// https://swagger.io/docs/specification/serialization/
func paramExplode(locatedIn ParameterLocation, explode *bool) bool {
	if explode != nil {
		return *explode
	}

	// When style is form, the default value is true.
	// For all other styles, the default value is false.
	if locatedIn == LocationQuery || locatedIn == LocationCookie {
		return true
	}

	return false
}
