package parser

import (
	"strings"

	"github.com/go-faster/errors"
	"golang.org/x/net/http/httpguts"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
)

func mergeParams(opParams, itemParams []*openapi.Parameter) []*openapi.Parameter {
	lookupOp := func(name string, in openapi.ParameterLocation) bool {
		for _, param := range opParams {
			if param.Name == name && param.In == in {
				return true
			}
		}
		return false
	}

	for _, param := range itemParams {
		// Param defined in operation take precedence over param defined in pathItem.
		if lookupOp(param.Name, param.In) {
			continue
		}

		opParams = append(opParams, param)
	}

	return opParams
}

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

		param, err := p.parseParameter(spec, newResolveCtx(p.depthLimit))
		if err != nil {
			return nil, errors.Wrapf(err, "parse parameter %q", spec.Name)
		}

		ploc := pnameLoc{
			name:     param.Name,
			location: param.In,
		}
		if _, ok := unique[ploc]; ok {
			err = errors.Errorf("duplicate parameter: %q in %q", param.Name, param.In)
			return nil, p.wrapLocation(spec, err)
		}

		unique[ploc] = struct{}{}
		result = append(result, param)
	}

	return result, nil
}

func validateParameter(name string, locatedIn openapi.ParameterLocation, param *ogen.Parameter) error {
	switch {
	case param.Schema != nil && param.Content != nil:
		return errors.New("parameter MUST contain either a schema property, or a content property, but not both")
	case param.Schema == nil && param.Content == nil:
		return errors.New("parameter MUST contain either a schema property, or a content property")
	case param.Content != nil && len(param.Content) < 1:
		// https://github.com/OAI/OpenAPI-Specification/discussions/2875
		return errors.New("content must have at least one entry")
	}

	// Path parameters are always required.
	switch locatedIn {
	case openapi.LocationPath:
		if !param.Required {
			return errors.New("path parameters must be required")
		}
	case openapi.LocationHeader:
		if !httpguts.ValidHeaderFieldName(name) {
			return errors.Errorf("invalid header name %q", name)
		}
	}
	return nil
}

func (p *parser) parseParameter(param *ogen.Parameter, ctx *resolveCtx) (_ *openapi.Parameter, rerr error) {
	if param == nil {
		return nil, errors.New("parameter object is empty or null")
	}
	defer func() {
		rerr = p.wrapLocation(param, rerr)
	}()
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

	if err := validateParameter(param.Name, locatedIn, param); err != nil {
		return nil, err
	}

	schema, err := p.parseSchema(param.Schema, ctx)
	if err != nil {
		return nil, errors.Wrap(err, "schema")
	}

	content, err := p.parseParameterContent(param.Content, ctx)
	if err != nil {
		return nil, errors.Wrap(err, "content")
	}

	op := &openapi.Parameter{
		Name:        param.Name,
		Description: param.Description,
		Schema:      schema,
		Content:     content,
		In:          locatedIn,
		Style:       inferParamStyle(locatedIn, param.Style),
		Explode:     inferParamExplode(locatedIn, param.Explode),
		Required:    param.Required,
		Deprecated:  param.Deprecated,
	}

	if param.Content != nil {
		// TODO: Validate content?
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
		return errors.Errorf("invalid in: %q", p.In)
	}

	types, ok := styles[stexp{p.Style, p.Explode}]
	if !ok {
		return errors.Errorf("invalid style explode combination %q, explode:%v", p.Style, p.Explode)
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
					return errors.New("all oneOf schemas must be simple types")
				}
			}

			if allowed(primitive) {
				return nil
			}
		}
	}

	return errors.Errorf("invalid schema:style:explode combination: (%q:%q:%v)",
		p.Schema.Type, p.Style, p.Explode)
}
