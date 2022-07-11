package jsonschema

import (
	ogenjson "github.com/ogen-go/ogen/json"
)

// LocationError is a wrapper for an error that has a location.
type LocationError = ogenjson.LocationError

func (p *Parser) wrapLocation(l ogenjson.Locatable, err error) error {
	if err == nil || l == nil || p == nil {
		return err
	}
	loc, ok := l.Location()
	if !ok {
		return err
	}
	filename := loc.Filename
	if filename == "" {
		filename = p.filename
	}
	return &LocationError{
		Loc: loc.WithFilename(filename),
		Err: err,
	}
}
