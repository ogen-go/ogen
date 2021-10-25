package parser

import (
	"strings"

	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen"
	ast "github.com/ogen-go/ogen/internal/ast2"
)

func (p *parser) parseParams(params []ogen.Parameter) ([]*ast.Parameter, error) {
	var result []*ast.Parameter
	for _, param := range params {
		param, err := p.parseParameter(param)
		if err != nil {
			return nil, xerrors.Errorf("parse parameter '%s': %w", param.Name, err)
		}

		result = append(result, param)
	}

	return result, nil
}

func (p *parser) parseParameter(param ogen.Parameter) (*ast.Parameter, error) {
	if ref := param.Ref; ref != "" {
		p, err := p.resolveParameter(ref)
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

	schema, err := p.parseSchema(param.Schema)
	if err != nil {
		return nil, xerrors.Errorf("schema: %w", err)
	}

	style, err := paramStyle(locatedIn, param.Style)
	if err != nil {
		return nil, xerrors.Errorf("style: %w", err)
	}

	return &ast.Parameter{
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
