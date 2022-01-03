package ir

import "github.com/ogen-go/ogen/internal/oas"

func (op *Operation) PathParams() []*Parameter   { return op.getParams(oas.LocationPath) }
func (op *Operation) QueryParams() []*Parameter  { return op.getParams(oas.LocationQuery) }
func (op *Operation) CookieParams() []*Parameter { return op.getParams(oas.LocationCookie) }
func (op *Operation) HeaderParams() []*Parameter { return op.getParams(oas.LocationHeader) }

func (op Operation) HasQueryParams() bool {
	for _, p := range op.Params {
		if p.Spec != nil && p.Spec.In.Query() {
			return true
		}
	}
	return false
}

func (op *Operation) getParams(locatedIn oas.ParameterLocation) []*Parameter {
	var params []*Parameter
	for _, p := range op.Params {
		if p.Spec.In == locatedIn {
			params = append(params, p)
		}
	}
	return params
}
