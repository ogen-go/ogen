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
		securitySchemes map[string]*ogen.SecuritySchema
	}

	external jsonschema.ExternalResolver
	schemas  map[string][]byte

	schemaParser *jsonschema.Parser
}

// Parse parses raw Spec into
func Parse(spec *ogen.Spec, s Settings) (*openapi.API, error) {
	spec.Init()
	p := &parser{
		spec:       spec,
		operations: nil,
		refs: struct {
			requestBodies   map[string]*openapi.RequestBody
			responses       map[string]*openapi.Response
			parameters      map[string]*openapi.Parameter
			headers         map[string]*openapi.Header
			examples        map[string]*openapi.Example
			securitySchemes map[string]*ogen.SecuritySchema
		}{
			requestBodies:   map[string]*openapi.RequestBody{},
			responses:       map[string]*openapi.Response{},
			parameters:      map[string]*openapi.Parameter{},
			headers:         map[string]*openapi.Header{},
			examples:        map[string]*openapi.Example{},
			securitySchemes: map[string]*ogen.SecuritySchema{},
		},
		external: s.External,
		schemas: map[string][]byte{
			"": spec.Raw,
		},
		schemaParser: jsonschema.NewParser(jsonschema.Settings{
			External: s.External,
			Resolver: componentsResolver{
				components: spec.Components.Schemas,
				root:       jsonschema.NewRootResolver(spec.Raw),
			},
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

	if err := p.parseOps(); err != nil {
		return nil, errors.Wrap(err, "parse operations")
	}

	return &openapi.API{
		Operations: p.operations,
		Components: components,
	}, nil
}

func (p *parser) parseOps() error {
	operationIDs := make(map[string]struct{})
	for path, item := range p.spec.Paths {
		if item == nil {
			return errors.Errorf("%s: pathItem object is empty or null", path)
		}
		if item.Ref != "" {
			return errors.Errorf("%s: referenced pathItem not supported", path)
		}

		itemParams, err := p.parseParams(item.Parameters)
		if err != nil {
			return errors.Wrapf(err, "%s: parameters", path)
		}

		if err := forEachOps(item, func(method string, op ogen.Operation) error {
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
		}); err != nil {
			return errors.Wrapf(err, "paths: %s", path)
		}
	}

	return nil
}

func (p *parser) parseOp(path, httpMethod string, spec ogen.Operation, itemParams []*openapi.Parameter) (_ *openapi.Operation, err error) {
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
		return nil, errors.Wrap(err, "parse path")
	}

	if spec.RequestBody != nil {
		op.RequestBody, err = p.parseRequestBody(spec.RequestBody, newResolveCtx())
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
