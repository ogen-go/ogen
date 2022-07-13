package location

import (
	"io"

	"github.com/go-faster/errors"
	yaml "github.com/go-faster/yamlx"
	"go.uber.org/multierr"
)

// PrintPrettyError prints the error in a pretty way and returns true if it was printed successfully.
func PrintPrettyError(w io.Writer, filename string, data []byte, err error) bool {
	// TODO(tdakkota): make it configurable?
	const (
		printLimit   = 5
		contextLines = 5
	)

	var lines Lines
	lines.Collect(data)

	write := func(msg string, loc Location, context int) {
		opts := PrintListingOptions{
			Filename: filename,
			Context:  contextLines,
		}
		_ = lines.PrintListing(w, msg, loc, opts)
	}

	if e, ok := errors.Into[*yaml.SyntaxError](err); ok {
		loc := Location{
			Line: e.Line,
		}
		write(e.Msg, loc, contextLines)
		return true
	}

	if e, ok := errors.Into[*yaml.TypeError](err); ok {
		printed := 0
		for _, e := range multierr.Errors(e.Group) {
			if printed >= printLimit {
				break
			}
			if e, ok := errors.Into[*yaml.UnmarshalError](e); ok && e.Node != nil {
				loc := Location{
					Line:   e.Node.Line,
					Column: e.Node.Column,
					Node:   e.Node,
				}
				write(e.Err.Error(), loc, contextLines)
				printed++
			}
		}
		// Consider the error as handled if it is printed at least once.
		return printed > 0
	}

	if e, ok := errors.Into[*Error](err); ok {
		var (
			iterErr = e.Err
			locErr  = e
		)
		for {
			e, ok := errors.Into[*Error](iterErr)
			if !ok {
				break
			}
			locErr = e
			iterErr = e.Err
		}
		// TODO(tdakkota): handle different files.
		if locErr.File == filename {
			write(locErr.Err.Error(), locErr.Loc, contextLines)
			return true
		}
	}

	return false
}
