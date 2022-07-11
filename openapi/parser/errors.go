package parser

import (
	ogenjson "github.com/ogen-go/ogen/json"
)

// LocationError is a wrapper for an error that has a location.
type LocationError = ogenjson.LocationError

func (p *parser) wrapLocation(file string, l ogenjson.Locatable, err error) error {
	if err == nil || l == nil || p == nil {
		return err
	}
	loc, ok := l.Location()
	if !ok {
		return err
	}
	filename := file
	if filename == "" {
		filename = p.filename
	}
	return &LocationError{
		Loc: loc.WithFilename(filename),
		Err: err,
	}
}
