package gen

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ogen-go/ogen"
)

func parseParameter(param ogen.Parameter, path string) (Parameter, error) {
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

	if t == "path" {
		exists, err := regexp.MatchString(fmt.Sprintf("{%s}", param.Name), path)
		if err != nil {
			return Parameter{}, fmt.Errorf("match path param '%s': %w", param.Name, err)
		}

		if !exists {
			return Parameter{}, fmt.Errorf("param '%s' not found in path '%s'", param.Name, path)
		}
	}

	var allowArrayType bool
	if t == LocationHeader {
		allowArrayType = true
	}

	pType, err := parseSimpleType(param.Schema, parseSimpleTypeParams{
		AllowArrays: allowArrayType,
	})
	if err != nil {
		return Parameter{}, fmt.Errorf("parse type: %w", err)
	}

	return Parameter{
		Name:       pascal(param.Name),
		In:         t,
		SourceName: param.Name,
		Type:       pType,
		Required:   param.Required,
	}, nil
}
