package ir

import (
	"github.com/ogen-go/ogen/internal/ast"
)

type Operation struct {
	Name      string
	PathParts []*PathPart
	Params    []*Parameter
	Request   *Request
	Response  *Response
	Spec      *ast.Operation
}

type PathPart struct {
	Raw   string
	Param *Parameter
}

type Parameter struct {
	Name string
	Type *Type
	Spec *ast.Parameter
}

type Request struct {
	Type     *Type
	Contents map[ContentType]*Type
	Required bool
	Spec     *ast.RequestBody
}

type Content struct {
	ContentType ContentType
	Type        *Type
}

// ContentType of body.
type ContentType string

// ContentTypeJSON is ContentType for json.
const ContentTypeJSON ContentType = "application/json"

func (t ContentType) JSON() bool { return t == ContentTypeJSON }

type Response struct {
	Type       *Type
	StatusCode map[int]*StatusResponse
	Default    *StatusResponse
	Spec       *ast.OperationResponse
}

type StatusResponse struct {
	NoContent *Type
	Contents  map[ContentType]*Type
	Spec      *ast.Response
}
