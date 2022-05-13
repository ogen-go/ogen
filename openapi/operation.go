package openapi

// Operation is an OpenAPI Operation.
type Operation struct {
	OperationID string // optional
	Description string // optional
	HTTPMethod  string
	Path        Path
	Parameters  []*Parameter
	RequestBody *RequestBody // optional

	// Security requirements.
	Security []SecurityRequirements

	// Operation responses.
	// Map is always non-nil.
	//
	// Key can be:
	//  * HTTP Status code
	//  * default
	//  * 1XX, 2XX, 3XX, 4XX, 5XX
	Responses map[string]*Response
}

type Path []PathPart

func (p Path) String() (path string) {
	for _, part := range p {
		if part.Raw != "" {
			path += part.Raw
			continue
		}
		path += "{" + part.Param.Name + "}"
	}
	return
}

// PathPart is a part of an OpenAPI Operation Path.
type PathPart struct {
	Raw   string
	Param *Parameter
}

// RequestBody of an OpenAPI Operation.
type RequestBody struct {
	Ref         string
	Description string
	Content     map[string]*MediaType
	Required    bool
}

// Response is an OpenAPI Response definition.
type Response struct {
	Ref         string
	Description string
	// Headers map[string]*Header
	Content map[string]*MediaType
	// Links map[string]*Link
}
