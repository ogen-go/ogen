// Package ogenerrors contains ogen errors type definitions and helpers.
package ogenerrors

import (
	"fmt"
	"net/http"
)

// Error is ogen error.
type Error interface {
	OperationID() string
	Code() int
	Unwrap() error
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

// Error implements error.
func (d *SecurityError) Error() string {
	return fmt.Sprintf("operation %s: security %q: %s", d.Operation, d.Security, d.Err)
}

// DecodeParamsError reports that error caused by params decoder.
type DecodeParamsError struct {
	Operation string
	Err       error
}

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

// Error implements error.
func (d *DecodeParamsError) Error() string {
	return fmt.Sprintf("operation %s: decode params: %s", d.Operation, d.Err)
}

// DecodeRequestError reports that error caused by request decoder.
type DecodeRequestError struct {
	Operation string
	Err       error
}

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

// Error implements error.
func (d *DecodeRequestError) Error() string {
	return fmt.Sprintf("operation %s: decode request: %s", d.Operation, d.Err)
}
