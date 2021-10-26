package ast

func CreateRequestBody() *RequestBody {
	return &RequestBody{
		Contents: map[string]*Schema{},
	}
}

func CreateOperationResponse() *OperationResponse {
	return &OperationResponse{
		StatusCode: map[int]*Response{},
	}
}

func CreateResponse() *Response {
	return &Response{Contents: map[string]*Schema{}}
}
