package location

import (
	"fmt"

	"github.com/go-faster/errors"
)

var _ interface {
	errors.Wrapper
	errors.Formatter
	fmt.Formatter
	error
} = (*Error)(nil)

// Error is a wrapper for an error that has a location.
type Error struct {
	File string
	Loc  Location
	Err  error
}

// Unwrap implements errors.Wrapper.
func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) fileName() string {
	filename := e.File
	if filename == "" || e.Loc.Line == 0 {
		return ""
	}
	return filename + ":"
}

// FormatError implements errors.Formatter.
func (e *Error) FormatError(p errors.Printer) (next error) {
	p.Printf("at %s%s", e.fileName(), e.Loc)
	return e.Err
}

// Format implements fmt.Formatter.
func (e *Error) Format(s fmt.State, verb rune) {
	errors.FormatError(e, s, verb)
}

// Error implements error.
func (e *Error) Error() string {
	return fmt.Sprintf("at %s%s: %s", e.fileName(), e.Loc, e.Err)
}
