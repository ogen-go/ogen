package ast

// ParameterLocation defines where OpenAPI parameter is located.
type ParameterLocation string

const (
	LocationQuery  ParameterLocation = "Query"
	LocationHeader ParameterLocation = "Header"
	LocationPath   ParameterLocation = "Path"
	LocationCookie ParameterLocation = "Cookie"
)

// Parameter is an OpenAPI Operation Parameter.
type Parameter struct {
	Name     string
	Schema   *Schema
	In       ParameterLocation
	Style    string
	Explode  bool
	Required bool
}
