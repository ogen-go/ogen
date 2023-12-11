package parser

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/location"
	"github.com/ogen-go/ogen/openapi"
)

// Expand generates an expanded ogen.Spec from given api.
func Expand(api *openapi.API) (*ogen.Spec, error) {
	e := expander{
		localToRemote: map[string]localToRemote{},
	}
	return e.Spec(api)
}

type localToRemote struct {
	ref jsonschema.Ref
	ptr location.Pointer
}

type expander struct {
	components    *ogen.Components
	localToRemote map[string]localToRemote
}

func (e *expander) Spec(api *openapi.API) (spec *ogen.Spec, err error) {
	spec = new(ogen.Spec)
	spec.Init()
	e.components = spec.Components

	spec.OpenAPI = api.Version.String()
	// FIXME(tdakkota): store actual information
	spec.Info = ogen.Info{
		Title:   "Expanded spec",
		Version: "v0.1.0",
	}

	if servers := api.Servers; len(servers) > 0 {
		expanded := make([]ogen.Server, len(servers))
		for i, s := range servers {
			expanded[i], err = e.Server(s)
			if err != nil {
				return nil, errors.Wrapf(err, "expand server [%d]", i)
			}
		}
		spec.Servers = expanded
	}

	setOperation := func(pi *ogen.PathItem, method string, op *ogen.Operation) error {
		var ptr **ogen.Operation
		switch strings.ToLower(method) {
		case "get":
			ptr = &pi.Get
		case "put":
			ptr = &pi.Put
		case "post":
			ptr = &pi.Post
		case "delete":
			ptr = &pi.Delete
		case "options":
			ptr = &pi.Options
		case "head":
			ptr = &pi.Head
		case "patch":
			ptr = &pi.Patch
		case "trace":
			ptr = &pi.Trace
		}
		if ptr == nil {
			return errors.Errorf("unexpected method %q", method)
		}

		if existing := *ptr; existing != nil {
			return errors.Errorf("path item already contains %q operation", method)
		}
		*ptr = op

		return nil
	}

	for _, op := range api.Operations {
		path := op.Path.String()

		expanded, err := e.Operation(op)
		if err != nil {
			return nil, errors.Wrapf(err, "expand operation %s %s", op.HTTPMethod, path)
		}

		pi := spec.Paths[path]
		if pi == nil {
			pi = &ogen.PathItem{}

			if spec.Paths == nil {
				spec.Paths = ogen.Paths{}
			}
			spec.Paths[path] = pi
		}

		if err := setOperation(pi, op.HTTPMethod, expanded); err != nil {
			return nil, err
		}
	}

	for _, wh := range api.Webhooks {
		for _, op := range wh.Operations {
			expanded, err := e.Operation(op)
			if err != nil {
				return nil, errors.Wrapf(err, "expand webhook operation %s %s", op.HTTPMethod, wh.Name)
			}

			pi := spec.Webhooks[wh.Name]
			if pi == nil {
				pi = &ogen.PathItem{}

				if spec.Webhooks == nil {
					spec.Webhooks = map[string]*ogen.PathItem{}
				}
				spec.Webhooks[wh.Name] = pi
			}

			if err := setOperation(pi, op.HTTPMethod, expanded); err != nil {
				return nil, err
			}
		}
	}

	return spec, nil
}

func (e *expander) Server(s openapi.Server) (expanded ogen.Server, err error) {
	expanded.Description = s.Description

	var (
		template strings.Builder
		vars     = map[string]ogen.ServerVariable{}
	)
	for _, part := range s.Template {
		if !part.IsParam() {
			template.WriteString(part.Raw)
			continue
		}
		param := part.Param

		template.WriteByte('{')
		template.WriteString(param.Name)
		template.WriteByte('}')

		vars[param.Name] = ogen.ServerVariable{
			Enum:        param.Enum,
			Default:     param.Default,
			Description: param.Description,
		}
	}
	expanded.URL = template.String()
	if len(vars) > 0 {
		expanded.Variables = vars
	}
	return expanded, nil
}

func (e *expander) SecurityRequirements(reqs openapi.SecurityRequirements) (expanded ogen.SecurityRequirements, err error) {
	if reqs == nil {
		return nil, nil
	}

	expanded = make(ogen.SecurityRequirements, len(reqs))
	for i, req := range reqs {
		expanded[i], err = e.SecurityRequirement(req)
		if err != nil {
			return nil, errors.Wrapf(err, "expand security requirement [%d]", i)
		}
	}
	return expanded, nil
}

func (e *expander) SecurityRequirement(req openapi.SecurityRequirement) (expanded ogen.SecurityRequirement, err error) {
	schemes := req.Schemes
	expanded = make(ogen.SecurityRequirement, len(schemes))
	for _, ss := range schemes {
		expanded[ss.Name] = ss.Scopes

		m := e.components.SecuritySchemes
		if _, ok := m[ss.Name]; !ok {
			m[ss.Name], err = e.SecurityScheme(ss.Security)
			if err != nil {
				return expanded, errors.Wrapf(err, "expand security scheme %q", ss.Name)
			}
		}
	}
	return expanded, nil
}

func (e *expander) SecurityScheme(s openapi.Security) (expanded *ogen.SecurityScheme, err error) {
	expanded = new(ogen.SecurityScheme)

	expanded.Type = s.Type
	expanded.Description = s.Description
	expanded.Name = s.Name
	expanded.In = s.In
	expanded.Scheme = s.Scheme
	expanded.BearerFormat = s.BearerFormat
	expanded.OpenIDConnectURL = s.OpenIDConnectURL

	expanded.Flows, err = e.OAuthFlows(&s.Flows)
	if err != nil {
		return nil, errors.Wrap(err, "expande oauth flows")
	}

	return expanded, nil
}

func (e *expander) OAuthFlows(flows *openapi.OAuthFlows) (expanded *ogen.OAuthFlows, err error) {
	if flows == nil {
		return nil, nil
	}
	expanded = new(ogen.OAuthFlows)

	expanded.Implicit, err = e.OAuthFlow(flows.Implicit)
	if err != nil {
		return nil, errors.Wrap(err, "expand implicit")
	}

	expanded.Password, err = e.OAuthFlow(flows.Password)
	if err != nil {
		return nil, errors.Wrap(err, "expand password")
	}

	expanded.ClientCredentials, err = e.OAuthFlow(flows.ClientCredentials)
	if err != nil {
		return nil, errors.Wrap(err, "expand clientCredentials")
	}

	expanded.AuthorizationCode, err = e.OAuthFlow(flows.AuthorizationCode)
	if err != nil {
		return nil, errors.Wrap(err, "expand authorizationCode")
	}

	return expanded, nil
}

func (e *expander) OAuthFlow(flow *openapi.OAuthFlow) (expanded *ogen.OAuthFlow, err error) {
	if flow == nil {
		return nil, nil
	}
	expanded = new(ogen.OAuthFlow)

	expanded.AuthorizationURL = flow.AuthorizationURL
	expanded.TokenURL = flow.TokenURL
	expanded.RefreshURL = flow.RefreshURL
	expanded.Scopes = flow.Scopes
	return expanded, nil
}

func (e *expander) Operation(op *openapi.Operation) (expanded *ogen.Operation, err error) {
	expanded = new(ogen.Operation)

	expanded.OperationID = op.OperationID
	expanded.Summary = op.Summary
	expanded.Description = op.Description
	expanded.Deprecated = op.Deprecated

	expanded.Security, err = e.SecurityRequirements(op.Security)
	if err != nil {
		return nil, errors.Wrap(err, "expand security")
	}

	expanded.Parameters, err = e.Parameters(op.Parameters)
	if err != nil {
		return nil, errors.Wrap(err, "expand parameters")
	}

	expanded.RequestBody, err = e.RequestBody(op.RequestBody)
	if err != nil {
		return nil, errors.Wrap(err, "expand requestBody")
	}

	expanded.Responses, err = e.Responses(op.Responses)
	if err != nil {
		return nil, errors.Wrap(err, "expand responses")
	}

	return expanded, nil
}

func (e *expander) Parameters(params []*openapi.Parameter) (expanded []*ogen.Parameter, err error) {
	if len(params) == 0 {
		return nil, nil
	}

	expanded = make([]*ogen.Parameter, len(params))
	for idx, param := range params {
		p, err := e.Parameter(param)
		if err != nil {
			return nil, errors.Wrapf(err, "expand param %d", idx)
		}
		expanded[idx] = p
	}

	return expanded, nil
}

func (e *expander) Parameter(param *openapi.Parameter) (expanded *ogen.Parameter, err error) {
	return e.genericParameter(param, "parameters", e.components.Parameters)
}

func (e *expander) genericParameter(
	param *openapi.Parameter,
	typ string,
	m map[string]*ogen.Parameter,
) (expanded *ogen.Parameter, err error) {
	expanded = new(ogen.Parameter)
	if ref := param.Ref; !ref.IsZero() {
		localRef, name, err := e.generateComponentLocalRef("#/components/"+typ+"/", ref, param.Pointer)
		if err != nil {
			return nil, err
		}

		ref := &ogen.Parameter{Ref: localRef}
		if _, ok := m[name]; !ok {
			m[name] = expanded
			defer func() {
				expanded = ref
			}()
		} else {
			return ref, nil
		}
	}

	expanded.Name = param.Name
	expanded.Description = param.Description
	expanded.Deprecated = param.Deprecated
	expanded.In = param.In.String()
	expanded.Style = param.Style.String()
	expanded.Explode = &param.Explode
	expanded.Required = param.Required
	expanded.AllowReserved = param.AllowReserved

	expanded.Schema, err = e.Schema(param.Schema, nil)
	if err != nil {
		return nil, errors.Wrap(err, "expand schema")
	}

	expanded.Content, err = e.ParameterContent(param.Content)
	if err != nil {
		return nil, errors.Wrap(err, "expand content")
	}

	return expanded, nil
}

func (e *expander) ParameterContent(content *openapi.ParameterContent) (map[string]ogen.Media, error) {
	if content == nil {
		return nil, nil
	}
	return e.Content(map[string]*openapi.MediaType{
		content.Name: content.Media,
	})
}

func (e *expander) RequestBody(body *openapi.RequestBody) (expanded *ogen.RequestBody, err error) {
	if body == nil {
		return nil, nil
	}

	expanded = new(ogen.RequestBody)
	if ref := body.Ref; !ref.IsZero() {
		localRef, name, err := e.generateComponentLocalRef("#/components/requestBodies/", ref, body.Pointer)
		if err != nil {
			return nil, err
		}

		ref := &ogen.RequestBody{Ref: localRef}
		m := e.components.RequestBodies
		if _, ok := m[name]; !ok {
			m[name] = expanded
			defer func() {
				expanded = ref
			}()
		} else {
			return ref, nil
		}
	}

	expanded.Description = body.Description
	expanded.Required = body.Required

	expanded.Content, err = e.Content(body.Content)
	if err != nil {
		return nil, errors.Wrap(err, "expand content")
	}

	return expanded, nil
}

func (e *expander) Responses(responses openapi.Responses) (expanded ogen.Responses, err error) {
	expanded = ogen.Responses{}

	for code, resp := range responses.StatusCode {
		pattern := strconv.Itoa(code)

		expanded[pattern], err = e.Response(resp)
		if err != nil {
			return nil, errors.Wrapf(err, "expand response %d", code)
		}
	}

	for idx, resp := range responses.Pattern {
		if resp == nil {
			continue
		}
		pattern := fmt.Sprintf("%dXX", idx+1)

		expanded[pattern], err = e.Response(resp)
		if err != nil {
			return nil, errors.Wrapf(err, "expand response %s", pattern)
		}
	}

	if resp := responses.Default; resp != nil {
		const pattern = "default"

		expanded[pattern], err = e.Response(resp)
		if err != nil {
			return nil, errors.Wrapf(err, "expand response %s", pattern)
		}
	}

	return expanded, nil
}

func (e *expander) Response(resp *openapi.Response) (expanded *ogen.Response, err error) {
	expanded = new(ogen.Response)
	if ref := resp.Ref; !ref.IsZero() {
		localRef, name, err := e.generateComponentLocalRef("#/components/responses/", ref, resp.Pointer)
		if err != nil {
			return nil, err
		}

		ref := &ogen.Response{Ref: localRef}
		m := e.components.Responses
		if _, ok := m[name]; !ok {
			m[name] = expanded
			defer func() {
				expanded = ref
			}()
		} else {
			return ref, nil
		}
	}

	expanded.Description = resp.Description

	expanded.Content, err = e.Content(resp.Content)
	if err != nil {
		return nil, errors.Wrap(err, "expand content")
	}

	expanded.Headers, err = e.Headers(resp.Headers)
	if err != nil {
		return nil, errors.Wrap(err, "expand headers")
	}

	return expanded, nil
}

func (e *expander) Headers(headers map[string]*openapi.Header) (expanded map[string]*ogen.Header, err error) {
	expanded = make(map[string]*ogen.Header, len(headers))

	for name, h := range headers {
		expanded[name], err = e.Header(h)
		if err != nil {
			return nil, errors.Wrapf(err, "expand header %q", name)
		}
	}

	return expanded, nil
}

func (e *expander) Header(h *openapi.Header) (expanded *ogen.Header, err error) {
	// Make a Paramater without "name" and "in".
	p := &openapi.Parameter{
		Ref: h.Ref,

		Description:   h.Description,
		Deprecated:    h.Deprecated,
		Schema:        h.Schema,
		Content:       h.Content,
		Style:         h.Style,
		Explode:       h.Explode,
		Required:      h.Required,
		AllowReserved: h.AllowReserved,
	}
	return e.genericParameter(p, "headers", e.components.Headers)
}

func (e *expander) Content(content map[string]*openapi.MediaType) (expanded map[string]ogen.Media, err error) {
	if content == nil {
		return nil, nil
	}
	expanded = make(map[string]ogen.Media, len(content))

	for name, media := range content {
		expanded[name], err = e.Media(media)
		if err != nil {
			return nil, errors.Wrapf(err, "expand media %q", name)
		}
	}

	return expanded, nil
}

func (e *expander) Media(media *openapi.MediaType) (expanded ogen.Media, err error) {
	expanded = ogen.Media{}

	expanded.Schema, err = e.Schema(media.Schema, nil)
	if err != nil {
		return expanded, errors.Wrap(err, "expand schema")
	}

	if encodings := media.Encoding; len(encodings) > 0 {
		expanded.Encoding = make(map[string]ogen.Encoding, len(encodings))
		for name, encoding := range encodings {
			expanded.Encoding[name], err = e.Encoding(encoding)
			if err != nil {
				return expanded, errors.Wrapf(err, "expand encoding %q", name)
			}
		}
	}

	return expanded, nil
}

func (e *expander) Encoding(media *openapi.Encoding) (expanded ogen.Encoding, err error) {
	expanded.ContentType = media.ContentType
	expanded.Style = media.Style.String()
	expanded.Explode = &media.Explode
	expanded.AllowReserved = media.AllowReserved

	expanded.Headers, err = e.Headers(media.Headers)
	if err != nil {
		return expanded, errors.Wrap(err, "expand headers")
	}
	return expanded, nil
}

func (e *expander) Schema(schema *jsonschema.Schema, walked map[*jsonschema.Schema]*ogen.Schema) (expanded *ogen.Schema, err error) {
	if schema == nil {
		return nil, nil
	}

	expanded = new(ogen.Schema)
	if ref := schema.Ref; !ref.IsZero() {
		localRef, name, err := e.generateComponentLocalRef("#/components/schemas/", ref, schema.Pointer)
		if err != nil {
			return nil, err
		}

		ref := &ogen.Schema{Ref: localRef}
		m := e.components.Schemas
		if _, ok := m[name]; !ok {
			m[name] = expanded
			defer func() {
				expanded = ref
			}()
		} else {
			return ref, nil
		}
	}
	if _, ok := walked[schema]; ok {
		return nil, errors.Errorf("recursive schema %q", schema.Ref)
	}
	if walked == nil {
		walked = map[*jsonschema.Schema]*ogen.Schema{}
	}
	walked[schema] = expanded
	defer func() {
		delete(walked, schema)
	}()

	expanded.Type = schema.Type.String()
	expanded.Format = schema.Format
	expanded.ContentEncoding = schema.ContentEncoding
	expanded.ContentMediaType = schema.ContentMediaType
	expanded.Summary = schema.Summary
	expanded.Description = schema.Description
	expanded.Deprecated = schema.Deprecated
	expanded.Nullable = schema.Nullable

	expanded.XML, err = e.XML(schema.XML)
	if err != nil {
		return nil, errors.Wrap(err, "expand xml")
	}

	expanded.Discriminator, err = e.Discriminator(schema.Discriminator)
	if err != nil {
		return nil, errors.Wrap(err, "expand discriminator")
	}

	expanded.AnyOf, err = e.Schemas(schema.AnyOf, walked)
	if err != nil {
		return nil, errors.Wrap(err, "expand anyOf")
	}

	expanded.OneOf, err = e.Schemas(schema.OneOf, walked)
	if err != nil {
		return nil, errors.Wrap(err, "expand oneOf")
	}

	expanded.AllOf, err = e.Schemas(schema.AllOf, walked)
	if err != nil {
		return nil, errors.Wrap(err, "expand allOf")
	}

	if enum := schema.Enum; len(enum) > 0 {
		expanded.Enum = make(jsonschema.Enum, len(enum))
		for i, e := range enum {
			raw, err := json.Marshal(e)
			if err != nil {
				return nil, errors.Wrapf(err, "marshal enum value [%d] %v", i, e)
			}
			expanded.Enum[i] = raw
		}
	}

	switch schema.Type {
	case jsonschema.Object:
		expanded.MinProperties = schema.MinProperties
		expanded.MaxProperties = schema.MaxProperties

		if props := schema.Properties; len(props) > 0 {
			expanded.Properties = make(ogen.Properties, len(props))
			for i, prop := range props {
				propSchema, err := e.Schema(prop.Schema, walked)
				if err != nil {
					return nil, errors.Wrapf(err, "expand property %q", prop.Name)
				}
				expanded.Properties[i] = ogen.Property{
					Name:   prop.Name,
					Schema: propSchema,
				}
				if prop.Required {
					expanded.Required = append(expanded.Required, prop.Name)
				}
			}
		}

		if ap := schema.AdditionalProperties; ap != nil {
			expandedAp := &ogen.AdditionalProperties{}

			if item := schema.Item; item != nil {
				s, err := e.Schema(item, walked)
				if err != nil {
					return nil, errors.Wrap(err, "expand additionalProperties")
				}
				if s != nil {
					expandedAp.Schema = *s
				}
			} else {
				v := new(bool)
				*v = *ap
				expandedAp.Bool = v
			}

			expanded.AdditionalProperties = expandedAp
		}

		if patternProps := schema.PatternProperties; len(patternProps) > 0 {
			expanded.PatternProperties = make(ogen.PatternProperties, len(patternProps))
			for i, prop := range patternProps {
				propSchema, err := e.Schema(prop.Schema, walked)
				if err != nil {
					return nil, errors.Wrapf(err, "expand pattern property %q", prop.Pattern)
				}
				expanded.Properties[i] = ogen.Property{
					Name:   prop.Pattern.String(),
					Schema: propSchema,
				}
			}
		}

	case jsonschema.Array:
		expanded.MinItems = schema.MinItems
		expanded.MaxItems = schema.MaxItems
		expanded.UniqueItems = schema.UniqueItems

		var item *ogen.Schema
		item, err = e.Schema(schema.Item, walked)
		if err != nil {
			return nil, errors.Wrap(err, "expand items")
		}
		expanded.Items = &ogen.Items{
			Item: item,
		}

	case jsonschema.Integer, jsonschema.Number:
		expanded.Minimum = schema.Minimum
		expanded.ExclusiveMinimum = schema.ExclusiveMinimum
		expanded.Maximum = schema.Maximum
		expanded.ExclusiveMaximum = schema.ExclusiveMaximum
		expanded.MultipleOf = schema.MultipleOf

	case jsonschema.String:
		expanded.MinLength = schema.MinLength
		expanded.MaxLength = schema.MaxLength
		expanded.Pattern = schema.Pattern

	case jsonschema.Boolean:
	case jsonschema.Null:
	}

	return expanded, nil
}

func (e *expander) Schemas(schemas []*jsonschema.Schema, walked map[*jsonschema.Schema]*ogen.Schema) (expanded []*ogen.Schema, err error) {
	if len(schemas) == 0 {
		return nil, nil
	}
	expanded = make([]*ogen.Schema, len(schemas))
	for i, s := range schemas {
		expanded[i], err = e.Schema(s, walked)
		if err != nil {
			return nil, errors.Wrapf(err, "expand %d", i)
		}
	}
	return expanded, nil
}

func (e *expander) Discriminator(d *jsonschema.Discriminator) (expanded *ogen.Discriminator, _ error) {
	if d == nil {
		return nil, nil
	}
	expanded = new(ogen.Discriminator)

	expanded.PropertyName = d.PropertyName
	if m := d.Mapping; len(m) > 0 {
		expanded.Mapping = make(map[string]string, len(m))
		for k, s := range m {
			if s.Ref.IsZero() {
				return nil, errors.Errorf("mapping %q has empty ref", k)
			}

			localRef, _, err := e.generateComponentLocalRef("#/components/schemas/", s.Ref, s.Pointer)
			if err != nil {
				return nil, errors.Wrapf(err, "mapping %q name", k)
			}

			expanded.Mapping[k] = localRef
		}
	}
	return expanded, nil
}

func (e *expander) XML(xml *jsonschema.XML) (expanded *ogen.XML, _ error) {
	if xml == nil {
		return nil, nil
	}
	expanded = new(ogen.XML)

	expanded.Name = xml.Name
	expanded.Namespace = xml.Namespace
	expanded.Prefix = xml.Prefix
	expanded.Attribute = xml.Attribute
	expanded.Wrapped = xml.Wrapped
	return expanded, nil
}

func (e *expander) generateComponentName(ref jsonschema.Ref) (string, error) {
	ptr := ref.Ptr
	idx := strings.LastIndexByte(ptr, '/')
	if idx < 0 || ptr[idx+1:] == "" {
		return "", errors.Errorf("can't generate component name for %q", ref)
	}
	return ptr[idx+1:], nil
}

func (e *expander) generateComponentLocalRef(
	prefix string,
	ref jsonschema.Ref,
	parentPtr location.Pointer,
) (localRef, name string, err error) {
	name, err = e.generateComponentName(ref)
	if err != nil {
		return "", "", err
	}

	localRef = prefix + name
	if existing, ok := e.localToRemote[localRef]; ok && existing.ref != ref {
		me := new(location.MultiError)
		me.ReportPtr(existing.ptr, fmt.Sprintf("local ref %q conflict: %q", localRef, existing.ref))
		me.ReportPtr(parentPtr, fmt.Sprintf("and %q", ref))
		return "", "", me
	}
	e.localToRemote[localRef] = localToRemote{
		ref: ref,
		ptr: parentPtr,
	}

	return localRef, name, nil
}
