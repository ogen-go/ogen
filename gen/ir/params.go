package ir

import (
	"slices"

	"github.com/ogen-go/ogen/openapi"
)

func (op *Operation) PathParams() []*Parameter   { return op.getParams(openapi.LocationPath) }
func (op *Operation) QueryParams() []*Parameter  { return op.getParams(openapi.LocationQuery) }
func (op *Operation) CookieParams() []*Parameter { return op.getParams(openapi.LocationCookie) }
func (op *Operation) HeaderParams() []*Parameter { return op.getParams(openapi.LocationHeader) }

func (op Operation) HasQueryParams() bool {
	return slices.ContainsFunc(op.Params, func(p *Parameter) bool {
		return p.Spec != nil && p.Spec.In.Query()
	})
}

func (op Operation) HasHeaderParams() bool {
	return slices.ContainsFunc(op.Params, func(p *Parameter) bool {
		return p.Spec != nil && p.Spec.In.Header()
	})
}

func (op Operation) HasCookieParams() bool {
	return slices.ContainsFunc(op.Params, func(p *Parameter) bool {
		return p.Spec != nil && p.Spec.In.Cookie()
	})
}

func (op Operation) PathParamsCount() (r int) {
	for _, p := range op.PathParts {
		if p.Param != nil {
			r++
		}
	}
	return r
}

func (op Operation) PathParamIndex(name string) int {
	idx := 0
	for _, p := range op.PathParts {
		if param := p.Param; param != nil {
			// Cut brackets '{', '}'.
			if n := param.Spec.Name; n == name {
				return idx
			}
			idx++
		}
	}
	return -1
}

func (op *Operation) getParams(locatedIn openapi.ParameterLocation) []*Parameter {
	var params []*Parameter
	for _, p := range op.Params {
		if p.Spec.In == locatedIn {
			params = append(params, p)
		}
	}
	return params
}
