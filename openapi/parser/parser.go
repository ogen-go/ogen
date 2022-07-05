package parser

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
)

type parser struct {
	// api spec, immutable.
	spec *ogen.Spec
	// parsed operations.
	operations []*openapi.Operation
	// refs contains lazy-initialized referenced components.
	refs struct {
		requestBodies   map[string]*openapi.RequestBody
		responses       map[string]*openapi.Response
		parameters      map[string]*openapi.Parameter
		headers         map[string]*openapi.Header
		examples        map[string]*openapi.Example
		securitySchemes map[string]*ogen.SecurityScheme
	}

	external   jsonschema.ExternalResolver
	schemas    map[string][]byte
	depthLimit int
	filename   string // optional, used for error messages

	schemaParser *jsonschema.Parser
}

// Parse parses raw Spec into
func Parse(spec *ogen.Spec, s Settings) (*openapi.API, error) {
	spec.Init()

	s.setDefaults()
	p := &parser{
		spec:       spec,
		operations: nil,
		refs: struct {
			requestBodies   map[string]*openapi.RequestBody
			responses       map[string]*openapi.Response
			parameters      map[string]*openapi.Parameter
			headers         map[string]*openapi.Header
			examples        map[string]*openapi.Example
			securitySchemes map[string]*ogen.SecurityScheme
		}{
			requestBodies:   map[string]*openapi.RequestBody{},
			responses:       map[string]*openapi.Response{},
			parameters:      map[string]*openapi.Parameter{},
			headers:         map[string]*openapi.Header{},
			examples:        map[string]*openapi.Example{},
			securitySchemes: map[string]*ogen.SecurityScheme{},
		},
		external: s.External,
		schemas: map[string][]byte{
			"": spec.Raw,
		},
		depthLimit: s.DepthLimit,
		filename:   s.Filename,
		schemaParser: jsonschema.NewParser(jsonschema.Settings{
			External: s.External,
			Resolver: componentsResolver{
				components: spec.Components.Schemas,
				root:       jsonschema.NewRootResolver(spec.Raw),
			},
			DepthLimit: s.DepthLimit,
			InferTypes: s.InferTypes,
		}),
	}

	for name, s := range spec.Components.SecuritySchemes {
		p.refs.securitySchemes[name] = s
	}

	components, err := p.parseComponents(spec.Components)
	if err != nil {
		return nil, errors.Wrap(err, "parse components")
	}

	if err := p.parsePathItems(); err != nil {
		return nil, errors.Wrap(err, "parse operations")
	}

	return &openapi.API{
		Operations: p.operations,
		Components: components,
	}, nil
}

func (p *parser) parsePathItems() error {
	operationIDs := make(map[string]struct{})
	for path, item := range p.spec.Paths {
		if err := p.parsePath(path, item, operationIDs); err != nil {
			return errors.Wrapf(err, "path %q", path)
		}
	}
	return nil
}

func (p *parser) parsePath(path string, item *ogen.PathItem, operationIDs map[string]struct{}) (rerr error) {
	if item == nil {
		return errors.New("pathItem object is empty or null")
	}
	if item.Ref != "" {
		return errors.New("referenced pathItem not supported")
	}
	defer func() {
		rerr = p.wrapLocation(item, rerr)
	}()

	itemParams, err := p.parseParams(item.Parameters)
	if err != nil {
		return errors.Wrap(err, "parameters")
	}

	return forEachOps(item, func(method string, op ogen.Operation) error {
		if id := op.OperationID; id != "" {
			if _, ok := operationIDs[id]; ok {
				return errors.Errorf("duplicate operationId: %q", id)
			}
			operationIDs[id] = struct{}{}
		}

		parsedOp, err := p.parseOp(path, method, op, itemParams)
		if err != nil {
			if op.OperationID != "" {
				return errors.Wrapf(err, "operation %q", op.OperationID)
			}
			return err
		}

		p.operations = append(p.operations, parsedOp)
		return nil
	})
}

func (p *parser) parseOp(
	path, httpMethod string,
	spec ogen.Operation,
	itemParams []*openapi.Parameter,
) (_ *openapi.Operation, err error) {
	defer func() {
		err = p.wrapLocation(&spec, err)
	}()

	op := &openapi.Operation{
		OperationID: spec.OperationID,
		Summary:     spec.Summary,
		Description: spec.Description,
		Deprecated:  spec.Deprecated,
		HTTPMethod:  httpMethod,
	}

	opParams, err := p.parseParams(spec.Parameters)
	if err != nil {
		return nil, errors.Wrap(err, "parameters")
	}

	// Merge operation parameters with pathItem parameters.
	op.Parameters = mergeParams(opParams, itemParams)

	op.Path, err = parsePath(path, op.Parameters)
	if err != nil {
		return nil, errors.Wrapf(err, "parse path %q", path)
	}

	if spec.RequestBody != nil {
		op.RequestBody, err = p.parseRequestBody(spec.RequestBody, newResolveCtx(p.depthLimit))
		if err != nil {
			return nil, errors.Wrap(err, "requestBody")
		}
	}

	op.Responses, err = p.parseResponses(spec.Responses)
	if err != nil {
		return nil, errors.Wrap(err, "responses")
	}

	if spec.Security != nil {
		// Use operation level security.
		op.Security, err = p.parseSecurityRequirements(spec.Security)
		if err != nil {
			return nil, errors.Wrap(err, "security")
		}
	} else {
		// Use root level security.
		op.Security, err = p.parseSecurityRequirements(p.spec.Security)
		if err != nil {
			return nil, errors.Wrap(err, "security")
		}
	}

	return op, nil
}

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
