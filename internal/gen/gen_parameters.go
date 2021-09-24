package gen

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ogen-go/ogen"
)

func (g *Generator) parseParameter(param ogen.Parameter, path string) (Parameter, error) {
	types := map[string]ParameterLocation{
		"query":  LocationQuery,
		"header": LocationHeader,
		"path":   LocationPath,
		"cookie": LocationCookie,
	}

	t, exists := types[strings.ToLower(param.In)]
	if !exists {
		return Parameter{}, fmt.Errorf("unsupported parameter type %s", param.In)
	}

	if t == LocationPath {
		if !param.Required {
			return Parameter{}, fmt.Errorf("parameters located in 'path' must be required")
		}

		exists, err := regexp.MatchString(fmt.Sprintf("{%s}", param.Name), path)
		if err != nil {
			return Parameter{}, fmt.Errorf("match path param '%s': %w", param.Name, err)
		}

		if !exists {
			return Parameter{}, fmt.Errorf("param '%s' not found in path '%s'", param.Name, path)
		}
	}

	name := pascal(param.Name)
	schema, err := g.generateSchema(name, param.Schema)
	if err != nil {
		return Parameter{}, fmt.Errorf("parse type: %w", err)
	}

	return Parameter{
		Name:        name,
		In:          t,
		SourceName:  param.Name,
		Type:        schema.typeName(),
		IsArrayType: param.Schema.Type == "array",
		Required:    param.Required,
	}, nil
}
