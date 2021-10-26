package ast

import "strings"

// ParameterLocation defines where OpenAPI parameter is located.
type ParameterLocation string

const (
	LocationQuery  ParameterLocation = "Query"
	LocationHeader ParameterLocation = "Header"
	LocationPath   ParameterLocation = "Path"
	LocationCookie ParameterLocation = "Cookie"
)

func (p ParameterLocation) Lower() string { return strings.ToLower(string(p)) }

// Parameter is an OpenAPI Operation Parameter.
type Parameter struct {
	Name     string
	Schema   *Schema
	In       ParameterLocation
	Style    string
	Explode  bool
	Required bool
}
