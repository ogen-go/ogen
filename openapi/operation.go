package openapi

import "github.com/ogen-go/ogen/internal/location"

// Operation is an OpenAPI Operation.
type Operation struct {
	OperationID string // optional
	Summary     string // optional
	Description string // optional
	Deprecated  bool   // optional

	HTTPMethod  string
	Path        Path
	Parameters  []*Parameter
	RequestBody *RequestBody // optional

	// Security requirements.
	Security []SecurityRequirement

	// Operation responses.
	// Map is always non-nil.
	//
	// Key can be:
	//  * HTTP Status code
	//  * default
	//  * 1XX, 2XX, 3XX, 4XX, 5XX
	Responses map[string]*Response

	location.Locator `json:"-" yaml:"-"`
}

// RequestBody of an OpenAPI Operation.
type RequestBody struct {
	Ref         string
	Description string
	Required    bool
	Content     map[string]*MediaType

	location.Locator `json:"-" yaml:"-"`
}

// Header is an OpenAPI Header definition.
type Header = Parameter

// Response is an OpenAPI Response definition.
type Response struct {
	Ref         string
	Description string
	Headers     map[string]*Header
	Content     map[string]*MediaType
	// Links map[string]*Link

	location.Locator `json:"-" yaml:"-"`
}
