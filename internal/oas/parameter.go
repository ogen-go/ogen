package oas

// ParameterLocation defines where OpenAPI parameter is located.
type ParameterLocation string

const (
	LocationQuery  ParameterLocation = "query"
	LocationHeader ParameterLocation = "header"
	LocationPath   ParameterLocation = "path"
	LocationCookie ParameterLocation = "cookie"
)

func (l ParameterLocation) Query() bool {
	return l == LocationQuery
}

func (l ParameterLocation) Header() bool {
	return l == LocationHeader
}

func (l ParameterLocation) Path() bool {
	return l == LocationPath
}

func (l ParameterLocation) Cookie() bool {
	return l == LocationCookie
}

// Parameter is an OpenAPI Operation Parameter.
type Parameter struct {
	Name        string
	Description string
	Schema      *Schema
	In          ParameterLocation
	Style       string
	Explode     bool
	Required    bool
}
