package ir

import ast "github.com/ogen-go/ogen/internal/ast2"

type Method struct {
	Name     string
	Params   []*Parameter
	Request  *Request
	Response *Response
	Spec     *ast.Method
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
	Spec       *ast.MethodResponse
}

type StatusResponse struct {
	NoContent *Type
	Contents  map[string]*Type
	Spec      *ast.Response
}
