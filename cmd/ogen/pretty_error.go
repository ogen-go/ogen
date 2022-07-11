package main

import (
	"fmt"
	"io"

	"github.com/go-faster/errors"
	"go.uber.org/multierr"
	"gopkg.in/yaml.v3"

	ogenjson "github.com/ogen-go/ogen/json"
)

func printPrettyError(w io.Writer, filename string, data []byte, err error) bool {
	var lines ogenjson.Lines
	lines.Collect(data)

	write := func(msg string, loc ogenjson.Location, context int) {
		_ = lines.PrettyError(w, filename, msg, loc, 5)
	}

	if e, ok := errors.Into[*yaml.SyntaxError](err); ok {
		loc := ogenjson.Location{
			Line: e.Line,
		}
		write(e.Msg, loc, 5)
		return true
	}

	if e, ok := errors.Into[*yaml.TypeError](err); ok {
		// TODO(tdakkota): make it configurable?
		const limit = 5
		printed := 0
		for _, e := range multierr.Errors(e.Group) {
			if printed >= limit {
				break
			}
			if e, ok := errors.Into[*yaml.UnmarshalError](e); ok && e.Node != nil {
				loc := ogenjson.Location{
					Line:   e.Node.Line,
					Column: e.Node.Column,
					Node:   e.Node,
				}
				write(e.Err.Error(), loc, 3)
				printed++
			}
		}
		// Consider the error as handled if it is printed at least once.
		return printed > 0
	}

	if e, ok := errors.Into[*ogenjson.LocationError](err); ok {
		var (
			iterErr = e.Err
			locErr  = e
		)
		for {
			e, ok := errors.Into[*ogenjson.LocationError](iterErr)
			if !ok {
				break
			}
			locErr = e
			iterErr = e.Err
		}
		// TODO(tdakkota): handle different files.
		if locErr.File == filename {
			write(locErr.Err.Error(), locErr.Loc, 5)
			return true
		}
	}

	return false
}

// PrettyError is pretty-printed error.
type PrettyError struct {
	Filename string
	Msg      string
	Loc      ogenjson.Location
	Lines    ogenjson.Lines
}

// Error implements error.
func (p *PrettyError) Error() string {
	return fmt.Sprintf("%s: %s", p.Loc, p.Msg)
}
