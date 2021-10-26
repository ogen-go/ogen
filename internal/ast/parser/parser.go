package parser

import (
	"strings"

	"github.com/ogen-go/ogen"
	ast "github.com/ogen-go/ogen/internal/ast"
	"golang.org/x/xerrors"
)

type parser struct {
	// schema specification, immutable.
	spec *ogen.Spec
	// parsed operations.
	operations []*ast.Operation
	// refs contains lazy-initialized referenced components.
	refs struct {
		schemas       map[string]*ast.Schema
		requestBodies map[string]*ast.RequestBody
		responses     map[string]*ast.Response
		parameters    map[string]*ast.Parameter
	}
}

func Parse(spec *ogen.Spec) ([]*ast.Operation, error) {
	spec.Init()
	p := &parser{
		spec: spec,
		refs: struct {
			schemas       map[string]*ast.Schema
			requestBodies map[string]*ast.RequestBody
			responses     map[string]*ast.Response
			parameters    map[string]*ast.Parameter
		}{
			schemas:       map[string]*ast.Schema{},
			requestBodies: map[string]*ast.RequestBody{},
			responses:     map[string]*ast.Response{},
			parameters:    map[string]*ast.Parameter{},
		},
	}

	err := p.parse()
	return p.operations, err
}

func (p *parser) parse() error {
	for path, item := range p.spec.Paths {
		if item.Ref != "" {
			return xerrors.New("referenced paths are not supported")
		}

		if err := forEachOps(item, func(method string, op ogen.Operation) error {
			return p.parseOp(path, strings.ToUpper(method), op)
		}); err != nil {
			return xerrors.Errorf("paths: %s: %w", path, err)
		}
	}

	return nil
}

func (p *parser) parseOp(path, httpMethod string, spec ogen.Operation) (err error) {
	op := &ast.Operation{
		OperationID: spec.OperationID,
		HTTPMethod:  httpMethod,
	}

	op.Parameters, err = p.parseParams(spec.Parameters)
	if err != nil {
		return xerrors.Errorf("parameters: %w", err)
	}

	op.PathParts, err = parsePath(path, op.Parameters)
	if err != nil {
		return xerrors.Errorf("parse path: %w", err)
	}

	if op.RequestBody != nil {
		op.RequestBody, err = p.parseRequestBody(spec.RequestBody)
		if err != nil {
			return xerrors.Errorf("requestBody: %w", err)
		}
	}

	if len(spec.Responses) > 0 {
		op.Responses, err = p.parseResponses(spec.Responses)
		if err != nil {
			return xerrors.Errorf("responses: %w", err)
		}
	}

	p.operations = append(p.operations, op)
	return nil
}
