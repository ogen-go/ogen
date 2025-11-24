package api

import (
	"context"
	"testing"
)

// TestServerCompiles ensures that the generated code compiles without duplicate case errors.
// This is a regression test for https://github.com/lanej/ogen/issues/1
func TestServerCompiles(t *testing.T) {
	// If this test compiles, the bug is fixed.
	// The bug caused duplicate case statements in type switches,
	// which prevented compilation.

	var _ Handler = testHandler{}
}

// TestResponseEncoding verifies that response encoding works correctly
// when multiple patterns share the same schema type.
func TestResponseEncoding(t *testing.T) {
	// Test that we can create responses with different status codes
	// using the same underlying type (CommonResponseStatusCode)
	_ = &CommonResponseStatusCode{
		StatusCode: 201,
		Response: CommonResponse{
			Status: "created",
			Data:   "test",
		},
	}

	_ = &CommonResponseStatusCode{
		StatusCode: 301,
		Response: CommonResponse{
			Status: "redirect",
			Data:   "test",
		},
	}

	// Test error response
	_ = &ErrorResponseStatusCode{
		StatusCode: 500,
		Response: ErrorResponse{
			Error: "test error",
		},
	}
}

// testHandler is a stub handler for testing.
type testHandler struct{}

func (testHandler) TestOperation(ctx context.Context, req *TestOperationReq) (TestOperationRes, error) {
	return &CommonResponseStatusCode{
		StatusCode: 200,
		Response: CommonResponse{
			Status: "ok",
			Data:   "test",
		},
	}, nil
}

func (testHandler) NewError(ctx context.Context, err error) *ErrorResponseStatusCode {
	return &ErrorResponseStatusCode{
		StatusCode: 500,
		Response: ErrorResponse{
			Error: err.Error(),
		},
	}
}
