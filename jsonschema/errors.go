package jsonschema

import (
	"github.com/ogen-go/ogen/internal/location"
)

// LocationError is a wrapper for an error that has a location.
type LocationError = location.Error

func (p *Parser) wrapLocation(filename string, l location.Locatable, err error) error {
	if err == nil || l == nil || p == nil {
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
