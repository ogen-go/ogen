package openapi

import "github.com/ogen-go/ogen/jsonschema"

// ParameterLocation defines where OpenAPI parameter is located.
type ParameterLocation string

const (
	LocationQuery  ParameterLocation = "query"
	LocationHeader ParameterLocation = "header"
	LocationPath   ParameterLocation = "path"
	LocationCookie ParameterLocation = "cookie"
)

func (l ParameterLocation) Query() bool { return l == LocationQuery }

func (l ParameterLocation) Header() bool { return l == LocationHeader }

func (l ParameterLocation) Path() bool { return l == LocationPath }

func (l ParameterLocation) Cookie() bool { return l == LocationCookie }

// Parameter is an OpenAPI Operation Parameter.
type Parameter struct {
	Ref         string
	Name        string
	Description string
	Schema      *jsonschema.Schema
	Content     map[string]*MediaType
	In          ParameterLocation
	Style       ParameterStyle
	Explode     bool
	Required    bool
}

type ParameterStyle string

func (s ParameterStyle) String() string { return string(s) }

// https://swagger.io/docs/specification/serialization/
const (
	PathStyleSimple ParameterStyle = "simple"
	PathStyleLabel  ParameterStyle = "label"
	PathStyleMatrix ParameterStyle = "matrix"

	QueryStyleForm           ParameterStyle = "form"
	QueryStyleSpaceDelimited ParameterStyle = "spaceDelimited"
	QueryStylePipeDelimited  ParameterStyle = "pipeDelimited"
	QueryStyleDeepObject     ParameterStyle = "deepObject"

	HeaderStyleSimple ParameterStyle = "simple"

	CookieStyleForm ParameterStyle = "form"
)
