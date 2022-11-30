package ogenerrors

import (
	"fmt"
	"net/http"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/openapi"
)

// DecodeParamsError reports that error caused by params decoder.
type DecodeParamsError struct {
	OperationContext
	Err error
}

// Code returns http code to respond.
func (d *DecodeParamsError) Code() int {
	return http.StatusBadRequest
}

// Unwrap returns child error.
func (d *DecodeParamsError) Unwrap() error {
	return d.Err
}

// FormatError implements errors.Formatter.
func (d *DecodeParamsError) FormatError(p errors.Printer) (next error) {
	p.Printf("operation %s: decode params", d.OperationName())
	return d.Err
}

// Format implements fmt.Formatter.
func (d *DecodeParamsError) Format(s fmt.State, verb rune) {
	errors.FormatError(d, s, verb)
}

// Error implements error.
func (d *DecodeParamsError) Error() string {
	return fmt.Sprintf("operation %s: decode params: %s", d.OperationName(), d.Err)
}

// DecodeParamError reports that error caused by parameter decoder.
type DecodeParamError struct {
	Name string
	In   openapi.ParameterLocation
	Err  error
}

// Unwrap returns child error.
func (d *DecodeParamError) Unwrap() error {
	return d.Err
}

// FormatError implements errors.Formatter.
func (d *DecodeParamError) FormatError(p errors.Printer) (next error) {
	p.Printf("%s: %q", d.In, d.Name)
	return d.Err
}

// Format implements fmt.Formatter.
func (d *DecodeParamError) Format(s fmt.State, verb rune) {
	errors.FormatError(d, s, verb)
}

// Error implements error.
func (d *DecodeParamError) Error() string {
	return fmt.Sprintf("%s: %q: %s", d.In, d.Name, d.Err)
}
