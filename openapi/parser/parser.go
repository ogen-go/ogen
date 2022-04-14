package parser

import (
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
	"github.com/ogen-go/ogen/openapi/document"
)

type parser struct {
	// api document, immutable.
	doc *document.Document
	// parsed operations.
	operations []*openapi.Operation
	// refs contains lazy-initialized referenced components.
	refs struct {
		requestBodies   map[string]*openapi.RequestBody
		responses       map[string]*openapi.Response
		parameters      map[string]*openapi.Parameter
		examples        map[string]*openapi.Example
		securitySchemes map[string]*document.SecuritySchema
	}

	schemaParser *jsonschema.Parser
}

func Parse(doc *document.Document, inferTypes bool) (*openapi.API, error) {
	doc.Init()
	p := &parser{
		doc:        doc,
		operations: nil,
		refs: struct {
			requestBodies   map[string]*openapi.RequestBody
			responses       map[string]*openapi.Response
			parameters      map[string]*openapi.Parameter
			examples        map[string]*openapi.Example
			securitySchemes map[string]*document.SecuritySchema
		}{
			requestBodies:   map[string]*openapi.RequestBody{},
			responses:       map[string]*openapi.Response{},
			parameters:      map[string]*openapi.Parameter{},
			examples:        map[string]*openapi.Example{},
			securitySchemes: map[string]*document.SecuritySchema{},
		},
		schemaParser: jsonschema.NewParser(jsonschema.Settings{
			Resolver: componentsResolver{
				components: doc.Components.Schemas,
				root:       jsonschema.NewRootResolver(doc.Raw),
			},
			InferTypes: inferTypes,
		}),
	}
	for name, s := range doc.Components.SecuritySchemes {
		p.refs.securitySchemes[name] = s
	}

	if err := p.parse(); err != nil {
		return nil, errors.Wrap(err, "parse document")
	}

	return &openapi.API{
		Operations: p.operations,
	}, nil
}

func (p *parser) parse() error {
	operationIDs := make(map[string]struct{})
	for path, item := range p.doc.Paths {
		if item == nil {
			return errors.Errorf("%s: unexpected nil schema", path)
		}
		if item.Ref != "" {
			return errors.Errorf("%s: referenced pathItem not supported", path)
		}

		itemParams, err := p.parseParams(item.Parameters)
		if err != nil {
			return errors.Wrapf(err, "%s: parameters", path)
		}

		if err := forEachOps(item, func(method string, op document.Operation) error {
			if id := op.OperationID; id != "" {
				if _, ok := operationIDs[id]; ok {
					return errors.Errorf("duplicate operationId: %q", id)
				}
				operationIDs[id] = struct{}{}
			}

			parsedOp, err := p.parseOp(path, method, op, itemParams)
			if err != nil {
				return errors.Wrapf(err, "operation %q", op.OperationID)
			}

			p.operations = append(p.operations, parsedOp)
			return nil
		}); err != nil {
			return errors.Wrapf(err, "paths: %s", path)
		}
	}

	return nil
}

func (p *parser) parseOp(path, httpMethod string, spec document.Operation, itemParams []*openapi.Parameter) (_ *openapi.Operation, err error) {
	op := &openapi.Operation{
		OperationID: spec.OperationID,
		Description: spec.Description,
		HTTPMethod:  strings.ToUpper(httpMethod),
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
		op.RequestBody, err = p.parseRequestBody(spec.RequestBody, resolveCtx{})
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
		op.Security, err = p.parseSecurityRequirements(p.doc.Security)
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
		// Param defined in operation take precedense over param defined in pathItem.
		if lookupOp(param.Name, param.In) {
			continue
		}

		opParams = append(opParams, param)
	}

	return opParams
}
