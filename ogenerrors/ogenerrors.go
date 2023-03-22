// Package ogenerrors contains ogen errors type definitions and helpers.
package ogenerrors

import (
	"fmt"
	"net/http"

	"github.com/go-faster/errors"
)

// Error is an ogen error.
type Error interface {
	OperationName() string
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

// OperationContext defines operation context for the error.
type OperationContext struct {
	Name string
	ID   string
}

// OperationName returns operation Name.
func (d OperationContext) OperationName() string {
	return d.Name
}

// OperationID returns operation ID.
func (d OperationContext) OperationID() string {
	return d.ID
}

// SecurityError reports that error caused by security handler.
type SecurityError struct {
	OperationContext
	Security string
	Err      error
}

// Code returns http code to respond.
func (d *SecurityError) Code() int {
	return http.StatusUnauthorized
}

// Unwrap returns child error.
func (d *SecurityError) Unwrap() error {
	return d.Err
}

// FormatError implements errors.Formatter.
func (d *SecurityError) FormatError(p errors.Printer) (next error) {
	p.Printf("operation %s: security %q", d.OperationName(), d.Security)
	return d.Err
}

// Format implements fmt.Formatter.
func (d *SecurityError) Format(s fmt.State, verb rune) {
	errors.FormatError(d, s, verb)
}

// Error implements error.
func (d *SecurityError) Error() string {
	return fmt.Sprintf("operation %s: security %q: %s", d.OperationName(), d.Security, d.Err)
}

// DecodeRequestError reports that error caused by request decoder.
type DecodeRequestError struct {
	OperationContext
	Err error
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
	p.Printf("operation %s: decode request", d.OperationName())
	return d.Err
}

// Format implements fmt.Formatter.
func (d *DecodeRequestError) Format(s fmt.State, verb rune) {
	errors.FormatError(d, s, verb)
}

// Error implements error.
func (d *DecodeRequestError) Error() string {
	return fmt.Sprintf("operation %s: decode request: %s", d.OperationName(), d.Err)
}
