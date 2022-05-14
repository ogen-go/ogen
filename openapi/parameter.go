package openapi

import "github.com/ogen-go/ogen/jsonschema"

// ParameterLocation defines where OpenAPI parameter is located.
type ParameterLocation string

const (
	// LocationQuery is "query" parameter location.
	LocationQuery ParameterLocation = "query"
	// LocationHeader is "header" parameter location.
	LocationHeader ParameterLocation = "header"
	// LocationPath is "path" parameter location.
	LocationPath ParameterLocation = "path"
	// LocationCookie is "cookie" parameter location.
	LocationCookie ParameterLocation = "cookie"
)

// Query whether parameter location is query.
func (l ParameterLocation) Query() bool { return l == LocationQuery }

// Header whether parameter location is header.
func (l ParameterLocation) Header() bool { return l == LocationHeader }

// Path whether parameter location is path.
func (l ParameterLocation) Path() bool { return l == LocationPath }

// Cookie whether parameter location is cookie.
func (l ParameterLocation) Cookie() bool { return l == LocationCookie }

// Parameter is an OpenAPI Operation Parameter.
type Parameter struct {
	Name        string
	Description string
	Schema      *jsonschema.Schema
	Content     map[string]*MediaType
	In          ParameterLocation
	Style       ParameterStyle
	Explode     bool
	Required    bool
}

// ParameterStyle is
type ParameterStyle string

// String implements fmt.Stringer.
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
