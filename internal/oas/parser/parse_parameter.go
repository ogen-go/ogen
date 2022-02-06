package parser

import (
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/oas"
)

func (p *parser) parseParams(params []*ogen.Parameter) ([]*oas.Parameter, error) {
	// Unique parameter is defined by a combination of a name and location.
	type pnameLoc struct {
		name     string
		location oas.ParameterLocation
	}

	var (
		result = make([]*oas.Parameter, 0, len(params))
		unique = make(map[pnameLoc]struct{})
	)

	for _, spec := range params {
		param, err := p.parseParameter(spec)
		if err != nil {
			return nil, errors.Wrapf(err, "parse parameter %q", spec.Name)
		}

		ploc := pnameLoc{
			name:     param.Name,
			location: param.In,
		}
		if _, ok := unique[ploc]; ok {
			return nil, errors.Errorf("duplicate parameter: %q in %q", param.Name, param.In)
		}

		unique[ploc] = struct{}{}
		result = append(result, param)
	}

	return result, nil
}

func (p *parser) parseParameter(param *ogen.Parameter) (*oas.Parameter, error) {
	if ref := param.Ref; ref != "" {
		parsed, err := p.resolveParameter(ref)
		if err != nil {
			return nil, errors.Wrapf(err, "resolve %q reference", ref)
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
		return nil, errors.Errorf("unsupported parameter type %s", param.In)
	}

	// Path parameters are always required.
	if locatedIn == oas.LocationPath && !param.Required {
		return nil, errors.New("path parameters must be required")
	}

	schema, err := p.schemaParser.Parse(param.Schema.ToJSONSchema())
	if err != nil {
		return nil, errors.Wrap(err, "schema")
	}

	style, err := paramStyle(locatedIn, param.Style)
	if err != nil {
		return nil, errors.Wrap(err, "style")
	}
	return &oas.Parameter{
		Name:        param.Name,
		Description: param.Description,
		In:          locatedIn,
		Schema:      schema,
		Style:       style,
		Explode:     paramExplode(locatedIn, param.Explode),
		Required:    param.Required,
	}, nil
}

// paramStyle checks parameter style field.
// https://swagger.io/docs/specification/serialization/
func paramStyle(locatedIn oas.ParameterLocation, style string) (oas.ParameterStyle, error) {
	if style == "" {
		defaultStyles := map[oas.ParameterLocation]oas.ParameterStyle{
			oas.LocationPath:   oas.PathStyleSimple,
			oas.LocationQuery:  oas.QueryStyleForm,
			oas.LocationHeader: oas.HeaderStyleSimple,
			oas.LocationCookie: oas.CookieStyleForm,
		}

		return defaultStyles[locatedIn], nil
	}

	allowedStyles := map[oas.ParameterLocation]map[string]oas.ParameterStyle{
		oas.LocationPath: {
			"simple": oas.PathStyleSimple,
			"label":  oas.PathStyleLabel,
			"matrix": oas.PathStyleMatrix,
		},
		oas.LocationQuery: {
			"form": oas.QueryStyleForm,
			// Not supported.
			// "spaceDelimited": struct{}{},
			"pipeDelimited": oas.QueryStylePipeDelimited,
			"deepObject":    oas.QueryStyleDeepObject,
		},
		oas.LocationHeader: {
			"simple": oas.HeaderStyleSimple,
		},
		oas.LocationCookie: {
			"form": oas.CookieStyleForm,
		},
	}

	s, found := allowedStyles[locatedIn][style]
	if !found {
		return "", errors.Errorf("unexpected style: %s", style)
	}

	return s, nil
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
