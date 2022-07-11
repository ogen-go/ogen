package json

import (
	"fmt"

	"github.com/go-faster/errors"
)

var _ interface {
	errors.Wrapper
	errors.Formatter
	fmt.Formatter
	error
} = (*LocationError)(nil)

// LocationError is a wrapper for an error that has a location.
type LocationError struct {
	File string
	Loc  Location
	Err  error
}

// Unwrap implements errors.Wrapper.
func (e *LocationError) Unwrap() error {
	return e.Err
}

func (e *LocationError) fileName() string {
	filename := e.File
	if filename == "" || e.Loc.Line == 0 {
		return ""
	}
	return filename + ":"
}

// FormatError implements errors.Formatter.
func (e *LocationError) FormatError(p errors.Printer) (next error) {
	p.Printf("at %s%s", e.fileName(), e.Loc)
	return e.Err
}

// Format implements fmt.Formatter.
func (e *LocationError) Format(s fmt.State, verb rune) {
	errors.FormatError(e, s, verb)
}

// Error implements error.
func (e *LocationError) Error() string {
	return fmt.Sprintf("at %s%s: %s", e.fileName(), e.Loc, e.Err)
}
