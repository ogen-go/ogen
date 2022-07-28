package parser

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/location"
)

// LocationError is a wrapper for an error that has a location.
type LocationError = location.Error

func (p *parser) wrapRef(filename string, l location.Locator, err error) error {
	if err == nil || p == nil {
		return err
	}
	return p.wrapLocation(filename, l.Field("$ref"), err)
}

func (p *parser) wrapField(field, filename string, l location.Locator, err error) error {
	if err == nil || p == nil {
		return err
	}
	return p.wrapLocation(filename, l.Field(field), err)
}

func (p *parser) wrapLocation(filename string, l location.Locator, err error) error {
	var locErr *LocationError
	// Do not wrap error if it is nil or is already a LocationError.
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
