package jsonschema

import (
	ogenjson "github.com/ogen-go/ogen/json"
)

// LocationError is a wrapper for an error that has a location.
type LocationError = ogenjson.LocationError

func (p *Parser) wrapLocation(filename string, l ogenjson.Locatable, err error) error {
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
