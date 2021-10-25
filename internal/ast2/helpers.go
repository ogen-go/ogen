package ast

func CreateRequestBody() *RequestBody {
	return &RequestBody{
		Contents: map[string]*Schema{},
	}
}

func CreateMethodResponse() *MethodResponse {
	return &MethodResponse{
		StatusCode: map[int]*Response{},
	}
}

func CreateResponse() *Response {
	return &Response{Contents: map[string]*Schema{}}
}
