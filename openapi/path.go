package openapi

// Path is an operation path.
type Path []PathPart[*Parameter]

// ID returns path, but without parameter names.
//
// For example, if path is "/users/{id}", ID returns "/users/{}".
func (p Path) ID() (path string) {
	for _, part := range p {
		if !part.IsParam() {
			path += part.Raw
			continue
		}
		path += "{}"
	}
	return
}

// String implements fmt.Stringer.
func (p Path) String() (path string) {
	for _, part := range p {
		if !part.IsParam() {
			path += part.Raw
			continue
		}
		path += "{" + part.Param.Name + "}"
	}
	return
}

// PathPart is a part of an OpenAPI Operation Path.
type PathPart[P any] struct {
	Raw   string
	Param P
}

// IsParam returns true if part is a parameter.
func (p PathPart[P]) IsParam() bool {
	return p.Raw == ""
}
