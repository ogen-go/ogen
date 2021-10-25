package parser

import "fmt"

type ErrPathParameterNotSpecified struct {
	ParamName string
}

func (e ErrPathParameterNotSpecified) Error() string {
	return fmt.Sprintf("path parameter '%s' not found in parameters", e.ParamName)
}

type ErrNotImplemented struct {
	Name string
}

func (e *ErrNotImplemented) Error() string {
	return e.Name + " not implemented"
}
