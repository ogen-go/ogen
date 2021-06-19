package gen

var (
	ParameterTypeQuery  ParameterType = "Query"
	ParameterTypeHeader ParameterType = "Header"
	ParameterTypePath   ParameterType = "Path"
	ParameterCookie     ParameterType = "Cookie"
)

type ParameterType string

type Parameter struct {
	Name string
	Type string
}
