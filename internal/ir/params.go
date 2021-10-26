package ir

import ast "github.com/ogen-go/ogen/internal/ast2"

func (op *Operation) PathParams() []*Parameter   { return op.getParams(ast.LocationPath) }
func (op *Operation) QueryParams() []*Parameter  { return op.getParams(ast.LocationQuery) }
func (op *Operation) CookieParams() []*Parameter { return op.getParams(ast.LocationCookie) }
func (op *Operation) HeaderParams() []*Parameter { return op.getParams(ast.LocationHeader) }

func (op *Operation) getParams(locatedIn ast.ParameterLocation) []*Parameter {
	var params []*Parameter
	for _, p := range op.Params {
		if p.Spec.In == locatedIn {
			params = append(params, p)
		}
	}
	return params
}
