package oas

// Operation is an OpenAPI Operation.
type Operation struct {
	OperationID string // optional
	HTTPMethod  string
	Path        Path
	Parameters  []*Parameter
	RequestBody *RequestBody // optional
	Responses   *OperationResponse
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
	Contents map[string]*Schema
	Required bool
}

// OperationResponse of an OpenAPI Operation.
type OperationResponse struct {
	StatusCode map[int]*Response
	Default    *Response
}

// Response is an OpenAPI Response definition.
type Response struct {
	Ref      string
	Contents map[string]*Schema
}
