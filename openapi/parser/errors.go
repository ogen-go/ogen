package parser

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/jsonpointer"
	"github.com/ogen-go/ogen/location"
)

// LocationError is a wrapper for an error that has a location.
type LocationError = location.Error

func (p *parser) file(ctx *jsonpointer.ResolveCtx) location.File {
	file := ctx.File()
	if file.IsZero() {
		return p.rootFile
	}
	return file
}

func (p *parser) wrapRef(file location.File, l location.Locator, err error) error {
	if err == nil || p == nil {
		return err
	}
	return p.wrapLocation(file, l.Field("$ref"), err)
}

func (p *parser) wrapField(field string, file location.File, l location.Locator, err error) error {
	if err == nil || p == nil {
		return err
	}
	return p.wrapLocation(file, l.Field(field), err)
}

func (p *parser) wrapLocation(file location.File, l location.Locator, err error) error {
	var locErr *LocationError
	// Do not wrap error if it is nil or is already a LocationError.
	if err == nil || p == nil || errors.As(err, &locErr) {
		return err
	}

	pos, ok := l.Position()
	if !ok {
		return err
	}
	if file.IsZero() {
		file = p.rootFile
	}
	return &LocationError{
		File: file,
		Pos:  pos,
		Err:  err,
	}
}
