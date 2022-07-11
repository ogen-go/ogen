package parser

import (
	ogenjson "github.com/ogen-go/ogen/json"
)

// LocationError is a wrapper for an error that has a location.
type LocationError = ogenjson.LocationError

func (p *parser) wrapLocation(filename string, l ogenjson.Locator, err error) error {
	if err == nil || p == nil {
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

func (p *parser) wrapField(field, filename string, l ogenjson.Locator, err error) error {
	if err == nil || p == nil {
		return err
	}
	return p.wrapLocation(filename, l.Field(field), err)
}
