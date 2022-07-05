package parser

import (
	"fmt"

	"github.com/go-faster/errors"

	ogenjson "github.com/ogen-go/ogen/json"
)

var _ interface {
	error
	errors.Wrapper
	errors.Formatter
} = (*LocationError)(nil)

// LocationError is a wrapper for an error that has a location.
type LocationError struct {
	file string
	loc  ogenjson.Location
	err  error
}

// Unwrap implements errors.Wrapper.
func (e *LocationError) Unwrap() error {
	return e.err
}

func (e *LocationError) fileName() string {
	filename := e.file
	if filename != "" {
		switch {
		case e.loc.Line != 0:
			// Line is set, so return "${filename}:".
			filename += ":"
		case e.loc.JSONPointer != "":
			// Line is not set, but JSONPointer is set, so return "${filename}#${JSONPointer}".
			filename += "#"
		default:
			// Neither line nor JSONPointer is set, so return empty string.
			return ""
		}
	}
	return filename
}

// FormatError implements errors.Formatter.
func (e *LocationError) FormatError(p errors.Printer) (next error) {
	p.Printf("at %s%s", e.fileName(), e.loc)
	return e.err
}

// Error implements error.
func (e *LocationError) Error() string {
	return fmt.Sprintf("at %s%s: %s", e.fileName(), e.loc, e.err)
}

func (p *parser) wrapLocation(l ogenjson.Locatable, err error) error {
	if err == nil {
		return nil
	}
	loc, ok := l.Location()
	if !ok {
		return err
	}
	return &LocationError{
		file: p.filename,
		loc:  loc,
		err:  err,
	}
}
