package parser

import (
	"strings"

	"github.com/ogen-go/ogen"
	ast "github.com/ogen-go/ogen/internal/ast2"
	"golang.org/x/xerrors"
)

type parser struct {
	spec *ogen.Spec
	refs struct {
		schemas       map[string]*ast.Schema
		requestBodies map[string]*ast.RequestBody
		responses     map[string]*ast.Response
		parameters    map[string]*ast.Parameter
	}
	methods []*ast.Method
}

func Parse(spec *ogen.Spec) ([]*ast.Method, error) {
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
	return p.methods, err
}

func (p *parser) parse() error {
	for path, item := range p.spec.Paths {
		if item.Ref != "" {
			return xerrors.New("referenced paths are not supported")
		}

		if err := forEachOps(item, func(method string, op ogen.Operation) error {
			return p.parseMethod(path, strings.ToUpper(method), op)
		}); err != nil {
			return xerrors.Errorf("paths: %s: %w", path, err)
		}
	}

	return nil
}

func (p *parser) parseMethod(path, httpMethod string, op ogen.Operation) (err error) {
	m := &ast.Method{
		OperationID: op.OperationID,
		HTTPMethod:  httpMethod,
	}

	m.Parameters, err = p.parseParams(op.Parameters)
	if err != nil {
		return xerrors.Errorf("parameters: %w", err)
	}

	m.PathParts, err = parsePath(path, m.PathParams())
	if err != nil {
		return xerrors.Errorf("parse path: %w", err)
	}

	if op.RequestBody != nil {
		m.RequestBody, err = p.parseRequestBody(op.RequestBody)
		if err != nil {
			return xerrors.Errorf("requestBody: %w", err)
		}
	}

	if len(op.Responses) > 0 {
		m.Responses, err = p.parseResponses(op.Responses)
		if err != nil {
			return xerrors.Errorf("responses: %w", err)
		}
	}

	p.methods = append(p.methods, m)
	return nil
}
