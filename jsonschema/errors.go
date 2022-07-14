package jsonschema

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/location"
)

// LocationError is a wrapper for an error that has a location.
type LocationError = location.Error

func (p *Parser) wrapLocation(filename string, l location.Locator, err error) error {
	var locErr *LocationError
	if err == nil || p == nil || errors.As(err, &locErr) {
		return err
	}
	loc, ok := l.Location()
	if !ok {
		return err
	}
	if filename == "" {
		filename = p.filename
	}
	return &LocationError{
		File: filename,
		Loc:  loc,
		Err:  err,
	}
}
