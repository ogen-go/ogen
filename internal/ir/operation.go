package ir

import ast "github.com/ogen-go/ogen/internal/ast"

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
	Contents map[string]*Type
	Required bool
	Spec     *ast.RequestBody
}

type Response struct {
	Type       *Type
	StatusCode map[int]*StatusResponse
	Default    *StatusResponse
	Spec       *ast.OperationResponse
}

type StatusResponse struct {
	NoContent *Type
	Contents  map[string]*Type
	Spec      *ast.Response
}
