package parser

import (
	"strings"

	"github.com/ogen-go/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/oas"
)

type parser struct {
	// schema specification, immutable.
	spec *ogen.Spec
	// parsed operations.
	operations []*oas.Operation
	// refs contains lazy-initialized referenced components.
	refs struct {
		schemas       map[string]*oas.Schema
		requestBodies map[string]*oas.RequestBody
		responses     map[string]*oas.Response
		parameters    map[string]*oas.Parameter
	}
}

func Parse(spec *ogen.Spec) ([]*oas.Operation, error) {
	spec.Init()
	p := &parser{
		spec: spec,
		refs: struct {
			schemas       map[string]*oas.Schema
			requestBodies map[string]*oas.RequestBody
			responses     map[string]*oas.Response
			parameters    map[string]*oas.Parameter
		}{
			schemas:       map[string]*oas.Schema{},
			requestBodies: map[string]*oas.RequestBody{},
			responses:     map[string]*oas.Response{},
			parameters:    map[string]*oas.Parameter{},
		},
	}

	err := p.parse()
	return p.operations, err
}

func (p *parser) parse() error {
	for path, item := range p.spec.Paths {
		if item.Ref != "" {
			return errors.Errorf("%s: referenced pathItem not supported", path)
		}

		if err := forEachOps(item, func(method string, op ogen.Operation) error {
			parsedOp, err := p.parseOp(path, strings.ToUpper(method), op, item.Parameters)
			if err != nil {
				return errors.Wrapf(err, "%s", strings.ToLower(method))
			}

			p.operations = append(p.operations, parsedOp)
			return nil
		}); err != nil {
			return errors.Wrapf(err, "paths: %s", path)
		}
	}

	return nil
}

func (p *parser) parseOp(path, httpMethod string, spec ogen.Operation, itemParameters []ogen.Parameter) (_ *oas.Operation, err error) {
	op := &oas.Operation{
		OperationID: spec.OperationID,
		HTTPMethod:  httpMethod,
	}

	op.Parameters, err = p.parseParams(mergeParams(spec.Parameters, itemParameters))
	if err != nil {
		return nil, errors.Wrap(err, "parameters")
	}

	op.PathParts, err = parsePath(path, op.Parameters)
	if err != nil {
		return nil, errors.Wrap(err, "parse path")
	}

	if spec.RequestBody != nil {
		op.RequestBody, err = p.parseRequestBody(spec.RequestBody)
		if err != nil {
			return nil, errors.Wrap(err, "requestBody")
		}
	}

	if len(spec.Responses) > 0 {
		op.Responses, err = p.parseResponses(spec.Responses)
		if err != nil {
			return nil, errors.Wrap(err, "responses")
		}
	}

	return op, nil
}

func mergeParams(opParams, itemParams []ogen.Parameter) []ogen.Parameter {
	if len(itemParams) == 0 {
		return opParams
	}

	lookupOp := func(name, in string) bool {
		for _, param := range opParams {
			if param.Name == name && param.In == in {
				return true
			}
		}
		return false
	}

	result := make([]ogen.Parameter, 0, len(opParams)+len(itemParams))
	result = append(result, opParams...)
	for _, param := range itemParams {
		// Param defined in operation take precedense over param defined in pathItem.
		if lookupOp(param.Name, param.In) {
			continue
		}

		result = append(result, param)
	}

	return result
}
