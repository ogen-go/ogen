package parser

import "github.com/ogen-go/ogen/internal/ast"

func createAstRBody() *ast.RequestBody {
	return &ast.RequestBody{
		Contents: map[string]*ast.Schema{},
	}
}

func createAstOpResponse() *ast.OperationResponse {
	return &ast.OperationResponse{
		StatusCode: map[int]*ast.Response{},
	}
}

func createAstResponse() *ast.Response {
	return &ast.Response{Contents: map[string]*ast.Schema{}}
}
