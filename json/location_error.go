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
	Loc Location
	Err error
}

// Unwrap implements errors.Wrapper.
func (e *LocationError) Unwrap() error {
	return e.Err
}

// FormatError implements errors.Formatter.
func (e *LocationError) FormatError(p errors.Printer) (next error) {
	p.Printf("at %s", e.Loc)
	return e.Err
}

// Format implements fmt.Formatter.
func (e *LocationError) Format(s fmt.State, verb rune) {
	errors.FormatError(e, s, verb)
}

// Error implements error.
func (e *LocationError) Error() string {
	return fmt.Sprintf("at %s: %s", e.Loc, e.Err)
}
