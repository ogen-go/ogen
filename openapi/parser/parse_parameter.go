package parser

import (
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseParams(params []*ogen.Parameter) ([]*openapi.Parameter, error) {
	// Unique parameter is defined by a combination of a name and location.
	type pnameLoc struct {
		name     string
		location openapi.ParameterLocation
	}

	var (
		result = make([]*openapi.Parameter, 0, len(params))
		unique = make(map[pnameLoc]struct{})
	)

	for idx, spec := range params {
		if spec == nil {
			return nil, errors.Errorf("parameter %d is empty or null", idx)
		}

		param, err := p.parseParameter(spec, resolveCtx{})
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

func (p *parser) parseParameter(param *ogen.Parameter, ctx resolveCtx) (*openapi.Parameter, error) {
	if param == nil {
		return nil, errors.New("parameter object is empty or null")
	}
	if ref := param.Ref; ref != "" {
		parsed, err := p.resolveParameter(ref, ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "resolve %q reference", ref)
		}
		return parsed, nil
	}

	types := map[string]openapi.ParameterLocation{
		"query":  openapi.LocationQuery,
		"header": openapi.LocationHeader,
		"path":   openapi.LocationPath,
		"cookie": openapi.LocationCookie,
	}

	locatedIn, exists := types[strings.ToLower(param.In)]
	if !exists {
		return nil, errors.Errorf("unsupported parameter type %q", param.In)
	}

	// Path parameters are always required.
	if locatedIn == openapi.LocationPath && !param.Required {
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
	return &openapi.Parameter{
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
func paramStyle(locatedIn openapi.ParameterLocation, style string) (openapi.ParameterStyle, error) {
	if style == "" {
		defaultStyles := map[openapi.ParameterLocation]openapi.ParameterStyle{
			openapi.LocationPath:   openapi.PathStyleSimple,
			openapi.LocationQuery:  openapi.QueryStyleForm,
			openapi.LocationHeader: openapi.HeaderStyleSimple,
			openapi.LocationCookie: openapi.CookieStyleForm,
		}

		return defaultStyles[locatedIn], nil
	}

	allowedStyles := map[openapi.ParameterLocation]map[string]openapi.ParameterStyle{
		openapi.LocationPath: {
			"simple": openapi.PathStyleSimple,
			"label":  openapi.PathStyleLabel,
			"matrix": openapi.PathStyleMatrix,
		},
		openapi.LocationQuery: {
			"form": openapi.QueryStyleForm,
			// Not supported.
			// "spaceDelimited": struct{}{},
			"pipeDelimited": openapi.QueryStylePipeDelimited,
			"deepObject":    openapi.QueryStyleDeepObject,
		},
		openapi.LocationHeader: {
			"simple": openapi.HeaderStyleSimple,
		},
		openapi.LocationCookie: {
			"form": openapi.CookieStyleForm,
		},
	}

	s, found := allowedStyles[locatedIn][style]
	if !found {
		return "", errors.Errorf("unexpected style: %q", style)
	}

	return s, nil
}

// paramExplode checks parameter explode field.
// https://swagger.io/docs/specification/serialization/
func paramExplode(locatedIn openapi.ParameterLocation, explode *bool) bool {
	if explode != nil {
		return *explode
	}

	// When style is form, the default value is true.
	// For all other styles, the default value is false.
	if locatedIn == openapi.LocationQuery || locatedIn == openapi.LocationCookie {
		return true
	}

	return false
}
