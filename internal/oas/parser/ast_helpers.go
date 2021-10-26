package parser

import "github.com/ogen-go/ogen/internal/oas"

func createAstRBody() *oas.RequestBody {
	return &oas.RequestBody{
		Contents: map[string]*oas.Schema{},
	}
}

func createAstOpResponse() *oas.OperationResponse {
	return &oas.OperationResponse{
		StatusCode: map[int]*oas.Response{},
	}
}

func createAstResponse() *oas.Response {
	return &oas.Response{Contents: map[string]*oas.Schema{}}
}
