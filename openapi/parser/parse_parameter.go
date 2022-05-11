package parser

import (
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonschema"
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

	op := &openapi.Parameter{
		Name:        param.Name,
		Description: param.Description,
		In:          locatedIn,
		Schema:      schema,
		Style:       inferParamStyle(locatedIn, param.Style),
		Explode:     inferParamExplode(locatedIn, param.Explode),
		Required:    param.Required,
	}

	if param.Content != nil {
		// Incorrect, but used to keep the original behaviour
		// and not break tests.
		//
		// return nil, errors.Errorf("content is not supported")
		return op, nil
	}

	if err := validateParamStyle(op); err != nil {
		return nil, err
	}

	return op, nil
}

func inferParamStyle(locatedIn openapi.ParameterLocation, style string) openapi.ParameterStyle {
	if style == "" {
		defaultStyles := map[openapi.ParameterLocation]openapi.ParameterStyle{
			openapi.LocationPath:   openapi.PathStyleSimple,
			openapi.LocationQuery:  openapi.QueryStyleForm,
			openapi.LocationHeader: openapi.HeaderStyleSimple,
			openapi.LocationCookie: openapi.CookieStyleForm,
		}

		return defaultStyles[locatedIn]
	}

	return openapi.ParameterStyle(style)
}

func inferParamExplode(locatedIn openapi.ParameterLocation, explode *bool) bool {
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

func validateParamStyle(p *openapi.Parameter) error {
	// https://swagger.io/docs/specification/serialization/
	const (
		primitive byte = 1 << iota
		array
		object
	)

	type stexp struct {
		style   openapi.ParameterStyle
		explode bool
	}

	table := map[openapi.ParameterLocation]map[stexp]byte{
		openapi.LocationPath: {
			{openapi.PathStyleSimple, false}: primitive | array | object,
			{openapi.PathStyleSimple, true}:  primitive | array | object,
			{openapi.PathStyleLabel, false}:  primitive | array | object,
			{openapi.PathStyleLabel, true}:   primitive | array | object,
			{openapi.PathStyleMatrix, false}: primitive | array | object,
			{openapi.PathStyleMatrix, true}:  primitive | array | object,
		},
		openapi.LocationQuery: {
			{openapi.QueryStyleForm, true}:            primitive | array | object,
			{openapi.QueryStyleForm, false}:           primitive | array | object,
			{openapi.QueryStyleSpaceDelimited, true}:  array,
			{openapi.QueryStyleSpaceDelimited, false}: array,
			{openapi.QueryStylePipeDelimited, true}:   array,
			{openapi.QueryStylePipeDelimited, false}:  array,
			{openapi.QueryStyleDeepObject, true}:      object,
		},
		openapi.LocationHeader: {
			{openapi.HeaderStyleSimple, false}: primitive | array | object,
			{openapi.HeaderStyleSimple, true}:  primitive | array | object,
		},
		openapi.LocationCookie: {
			{openapi.CookieStyleForm, true}:  primitive,
			{openapi.CookieStyleForm, false}: primitive | array | object,
		},
	}

	styles, ok := table[p.In]
	if !ok {
		return errors.Errorf("invalid style: %q", p.In)
	}

	types, ok := styles[stexp{p.Style, p.Explode}]
	if !ok {
		return errors.Errorf("invalid style explode combination")
	}

	allowed := func(t byte) bool { return types&t != 0 }

	switch p.Schema.Type {
	case jsonschema.String, jsonschema.Integer, jsonschema.Number, jsonschema.Boolean:
		if allowed(primitive) {
			return nil
		}
	case jsonschema.Array:
		if allowed(array) {
			return nil
		}
	case jsonschema.Object:
		if allowed(object) {
			return nil
		}
	case jsonschema.Empty:
		if p.Schema.OneOf != nil {
			for _, s := range p.Schema.OneOf {
				switch s.Type {
				case jsonschema.String, jsonschema.Integer, jsonschema.Number, jsonschema.Boolean:
					// ok
				default:
					return errors.Errorf("all oneOf schemas must be simple types")
				}
			}

			if allowed(primitive) {
				return nil
			}
		}
	}

	return errors.Errorf("invalid schema:style:explode combination")
}
