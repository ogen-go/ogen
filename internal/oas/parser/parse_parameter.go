package parser

import (
	"strings"

	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/oas"
)

func (p *parser) parseParams(params []ogen.Parameter) ([]*oas.Parameter, error) {
	var result []*oas.Parameter
	for _, param := range params {
		parsed, err := p.parseParameter(param)
		if err != nil {
			return nil, xerrors.Errorf("parse parameter '%s': %w", param.Name, err)
		}

		result = append(result, parsed)
	}

	return result, nil
}

func (p *parser) parseParameter(param ogen.Parameter) (*oas.Parameter, error) {
	if ref := param.Ref; ref != "" {
		parsed, err := p.resolveParameter(ref)
		if err != nil {
			return nil, xerrors.Errorf("resolve '%s' reference: %w", ref, err)
		}
		return parsed, nil
	}

	types := map[string]oas.ParameterLocation{
		"query":  oas.LocationQuery,
		"header": oas.LocationHeader,
		"path":   oas.LocationPath,
		"cookie": oas.LocationCookie,
	}

	locatedIn, exists := types[strings.ToLower(param.In)]
	if !exists {
		return nil, xerrors.Errorf("unsupported parameter type %s", param.In)
	}

	schema, err := p.parseSchema(param.Schema)
	if err != nil {
		return nil, xerrors.Errorf("schema: %w", err)
	}

	style, err := paramStyle(locatedIn, param.Style)
	if err != nil {
		return nil, xerrors.Errorf("style: %w", err)
	}

	return &oas.Parameter{
		Name:     param.Name,
		In:       locatedIn,
		Schema:   schema,
		Style:    style,
		Explode:  paramExplode(locatedIn, param.Explode),
		Required: param.Required,
	}, nil
}

// paramStyle checks parameter style field.
// https://swagger.io/docs/specification/serialization/
func paramStyle(locatedIn oas.ParameterLocation, style string) (string, error) {
	if style == "" {
		defaultStyles := map[oas.ParameterLocation]string{
			oas.LocationPath:   "simple",
			oas.LocationQuery:  "form",
			oas.LocationHeader: "simple",
			oas.LocationCookie: "form",
		}

		return defaultStyles[locatedIn], nil
	}

	allowedStyles := map[oas.ParameterLocation]map[string]struct{}{
		oas.LocationPath: {
			"simple": struct{}{},
			"label":  struct{}{},
			"matrix": struct{}{},
		},
		oas.LocationQuery: {
			"form": struct{}{},
			// Not supported.
			// "spaceDelimited": struct{}{},
			"pipeDelimited": struct{}{},
			"deepObject":    struct{}{},
		},
		oas.LocationHeader: {
			"simple": struct{}{},
		},
		oas.LocationCookie: {
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
func paramExplode(locatedIn oas.ParameterLocation, explode *bool) bool {
	if explode != nil {
		return *explode
	}

	// When style is form, the default value is true.
	// For all other styles, the default value is false.
	if locatedIn == oas.LocationQuery || locatedIn == oas.LocationCookie {
		return true
	}

	return false
}
