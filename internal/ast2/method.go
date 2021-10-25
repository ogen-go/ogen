package ast

type Method struct {
	OperationID string
	HTTPMethod  string
	PathParts   []PathPart
	Parameters  []*Parameter
	RequestBody *RequestBody
	Responses   *MethodResponse
}

type PathPart struct {
	Raw   string
	Param *Parameter
}

type RequestBody struct {
	Contents map[string]*Schema
	Required bool
}

type MethodResponse struct {
	StatusCode map[int]*Response
	Default    *Response
}

type Response struct {
	Contents map[string]*Schema
}

func (m *Method) Path() (path string) {
	for _, part := range m.PathParts {
		if part.Raw != "" {
			path += part.Raw
			continue
		}
		path += "{" + part.Param.Name + "}"
	}
	return
}

func (m *Method) PathParams() []*Parameter   { return m.getParams(LocationPath) }
func (m *Method) QueryParams() []*Parameter  { return m.getParams(LocationQuery) }
func (m *Method) CookieParams() []*Parameter { return m.getParams(LocationCookie) }
func (m *Method) HeaderParams() []*Parameter { return m.getParams(LocationHeader) }

func (m *Method) getParams(locatedIn ParameterLocation) []*Parameter {
	var params []*Parameter
	for _, p := range m.Parameters {
		if p.In == locatedIn {
			params = append(params, p)
		}
	}
	return params
}
