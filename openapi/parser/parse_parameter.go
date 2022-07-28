package parser

import (
	"strings"

	"github.com/go-faster/errors"
	"golang.org/x/net/http/httpguts"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/location"
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

func (p *parser) parseParams(
	params []*ogen.Parameter,
	locator location.Locator,
	ctx *resolveCtx,
) ([]*openapi.Parameter, error) {
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
			loc := locator.Index(idx)
			err := errors.Errorf("parameter %d is empty or null", idx)
			return nil, p.wrapLocation(ctx.lastLoc(), loc, err)
		}

		param, err := p.parseParameter(spec, ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "parse parameter %q", spec.Name)
		}

		ploc := pnameLoc{
			name:     param.Name,
			location: param.In,
		}
		if _, ok := unique[ploc]; ok {
			err := errors.Errorf("duplicate parameter: %q in %q", param.Name, param.In)
			return nil, p.wrapLocation(ctx.lastLoc(), spec.Locator, err)
		}

		unique[ploc] = struct{}{}
		result = append(result, param)
	}

	return result, nil
}

func (p *parser) validateParameter(
	name string,
	locatedIn openapi.ParameterLocation,
	param *ogen.Parameter,
	file string,
) error {
	switch {
	case param.Schema != nil && param.Content != nil:
		err := errors.New("parameter MUST contain either a schema property, or a content property, but not both")
		return p.wrapField("schema", file, param.Locator, err)
	case param.Schema == nil && param.Content == nil:
		return errors.New("parameter MUST contain either a schema property, or a content property")
	case param.Content != nil && len(param.Content) < 1:
		// https://github.com/OAI/OpenAPI-Specification/discussions/2875
		err := errors.New("content must have at least one entry")
		return p.wrapField("content", file, param.Locator, err)
	}

	// Path parameters are always required.
	switch locatedIn {
	case openapi.LocationPath:
		if !param.Required {
			err := errors.New("path parameters must be required")
			return p.wrapField("required", file, param.Locator, err)
		}
	case openapi.LocationHeader:
		if !httpguts.ValidHeaderFieldName(name) {
			err := errors.Errorf("invalid header name %q", name)
			return p.wrapField("name", file, param.Locator, err)
		}
	}
	return nil
}

func (p *parser) parseParameter(param *ogen.Parameter, ctx *resolveCtx) (_ *openapi.Parameter, rerr error) {
	if param == nil {
		return nil, errors.New("parameter object is empty or null")
	}
	defer func() {
		rerr = p.wrapLocation(ctx.lastLoc(), param.Locator, rerr)
	}()
	if ref := param.Ref; ref != "" {
		parsed, err := p.resolveParameter(ref, ctx)
		if err != nil {
			return nil, p.wrapRef(ctx.lastLoc(), param.Locator, err)
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
		err := errors.Errorf("unknown parameter location %q", param.In)
		return nil, p.wrapField("in", ctx.lastLoc(), param.Locator, err)
	}

	if err := p.validateParameter(param.Name, locatedIn, param, ctx.lastLoc()); err != nil {
		return nil, err
	}

	schema, err := p.parseSchema(param.Schema, ctx)
	if err != nil {
		err := errors.Wrap(err, "schema")
		return nil, p.wrapField("schema", ctx.lastLoc(), param.Locator, err)
	}

	content, err := p.parseParameterContent(param.Content, param.Locator.Field("content"), ctx)
	if err != nil {
		err := errors.Wrap(err, "content")
		return nil, p.wrapField("content", ctx.lastLoc(), param.Locator, err)
	}

	op := &openapi.Parameter{
		Name:          param.Name,
		Description:   param.Description,
		Deprecated:    param.Deprecated,
		Schema:        schema,
		Content:       content,
		In:            locatedIn,
		Style:         inferParamStyle(locatedIn, param.Style),
		Explode:       inferParamExplode(locatedIn, param.Explode),
		Required:      param.Required,
		AllowReserved: param.AllowReserved,
		Locator:       param.Locator,
	}

	// TODO: Validate content?
	if param.Content == nil {
		if err := p.validateParamStyle(op, ctx.lastLoc()); err != nil {
			return nil, err
		}
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

func (p *parser) validateParamStyle(param *openapi.Parameter, loc string) error {
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
	wrap := func(field string, err error) error {
		return p.wrapField(field, loc, param.Locator, err)
	}

	styles, ok := table[param.In]
	if !ok {
		return wrap("in", errors.Errorf(`invalid "in": %q`, param.In))
	}

	types, ok := styles[stexp{param.Style, param.Explode}]
	if !ok {
		err := errors.Errorf("invalid style explode combination %q, explode:%v", param.Style, param.Explode)
		return wrap("style", err)
	}

	allowed := func(t byte) bool { return types&t != 0 }

	switch param.Schema.Type {
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
		if param.Schema.OneOf != nil {
			for _, s := range param.Schema.OneOf {
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
		param.Schema.Type, param.Style, param.Explode)
}
