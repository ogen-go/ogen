package ogenreflect

import "github.com/ogen-go/ogen/openapi"

// ParameterKey is a map key for parameters.
type ParameterKey struct {
	// Name is the name of the parameter.
	Name string
	// In is the location of the parameter.
	In openapi.ParameterLocation
}

// ParameterMap is a generic map of parameters.
type ParameterMap[V any] map[ParameterKey]V

func (p ParameterMap[V]) find(name string, in openapi.ParameterLocation) (v V, ok bool) {
	v, ok = p[ParameterKey{Name: name, In: in}]
	return v, ok
}

// Query returns a parameter from the query.
func (p ParameterMap[V]) Query(name string) (V, bool) {
	return p.find(name, openapi.LocationQuery)
}

// Header returns a parameter from the header.
func (p ParameterMap[V]) Header(name string) (V, bool) {
	return p.find(name, openapi.LocationHeader)
}

// Path returns a parameter from the path.
func (p ParameterMap[V]) Path(name string) (V, bool) {
	return p.find(name, openapi.LocationPath)
}

// Cookie returns a parameter from the cookie.
func (p ParameterMap[V]) Cookie(name string) (V, bool) {
	return p.find(name, openapi.LocationCookie)
}
