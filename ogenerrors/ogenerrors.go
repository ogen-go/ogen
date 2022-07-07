// Package ogenerrors contains ogen errors type definitions and helpers.
package ogenerrors

import (
	"fmt"
	"net/http"

	"github.com/go-faster/errors"
)

// Error is ogen error.
type Error interface {
	OperationID() string
	Code() int
	errors.Wrapper
	errors.Formatter
	fmt.Formatter
	error
}

var _ = []Error{
	new(SecurityError),
	new(DecodeParamsError),
	new(DecodeRequestError),
}

// SecurityError reports that error caused by security handler.
type SecurityError struct {
	Operation string
	Security  string
	Err       error
}

// OperationID returns operation ID of failed request.
func (d *SecurityError) OperationID() string {
	return d.Operation
}

// Code returns http code to respond.
func (d *SecurityError) Code() int {
	return http.StatusBadRequest
}

// Unwrap returns child error.
func (d *SecurityError) Unwrap() error {
	return d.Err
}

// FormatError implements errors.Formatter.
func (d *SecurityError) FormatError(p errors.Printer) (next error) {
	p.Printf("operation %s: security %q", d.Operation, d.Security)
	return d.Err
}

// Format implements fmt.Formatter.
func (d *SecurityError) Format(s fmt.State, verb rune) {
	errors.FormatError(d, s, verb)
}

// Error implements error.
func (d *SecurityError) Error() string {
	return fmt.Sprintf("operation %s: security %q: %s", d.Operation, d.Security, d.Err)
}

// DecodeParamsError reports that error caused by params decoder.
type DecodeParamsError struct {
	Operation string
	Err       error
}

// OperationID returns operation ID of failed request.
func (d *DecodeParamsError) OperationID() string {
	return d.Operation
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
	p.Printf("operation %s: decode params", d.Operation)
	return d.Err
}

// Format implements fmt.Formatter.
func (d *DecodeParamsError) Format(s fmt.State, verb rune) {
	errors.FormatError(d, s, verb)
}

// Error implements error.
func (d *DecodeParamsError) Error() string {
	return fmt.Sprintf("operation %s: decode params: %s", d.Operation, d.Err)
}

// DecodeRequestError reports that error caused by request decoder.
type DecodeRequestError struct {
	Operation string
	Err       error
}

// OperationID returns operation ID of failed request.
func (d *DecodeRequestError) OperationID() string {
	return d.Operation
}

// Code returns http code to respond.
func (d *DecodeRequestError) Code() int {
	return http.StatusBadRequest
}

// Unwrap returns child error.
func (d *DecodeRequestError) Unwrap() error {
	return d.Err
}

// FormatError implements errors.Formatter.
func (d *DecodeRequestError) FormatError(p errors.Printer) (next error) {
	p.Printf("operation %s: decode request", d.Operation)
	return d.Err
}

// Format implements fmt.Formatter.
func (d *DecodeRequestError) Format(s fmt.State, verb rune) {
	errors.FormatError(d, s, verb)
}

// Error implements error.
func (d *DecodeRequestError) Error() string {
	return fmt.Sprintf("operation %s: decode request: %s", d.Operation, d.Err)
}
