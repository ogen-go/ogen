package location

import (
	"fmt"
	"io"

	"github.com/go-faster/errors"
	"github.com/go-faster/yaml"
	"go.uber.org/multierr"
)

func firstNonEmpty(strs ...string) string {
	for _, s := range strs {
		if s != "" {
			return s
		}
	}
	return ""
}

var _ interface {
	errors.Wrapper
	errors.Formatter
	fmt.Formatter
	error
} = (*Error)(nil)

// Error is a wrapper for an error that has a location.
type Error struct {
	File File
	Pos  Position
	Err  error
}

// Unwrap implements errors.Wrapper.
func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) fileName() string {
	filename := firstNonEmpty(e.File.Name, e.File.Source)
	if filename == "" || e.Pos.Line == 0 {
		return ""
	}
	return filename + ":"
}

// FormatError implements errors.Formatter.
func (e *Error) FormatError(p errors.Printer) error {
	p.Printf("at %s%s", e.fileName(), e.Pos)
	return e.Err
}

// Format implements fmt.Formatter.
func (e *Error) Format(s fmt.State, verb rune) {
	errors.FormatError(e, s, verb)
}

// Error implements error.
func (e *Error) Error() string {
	return fmt.Sprintf("at %s%s: %s", e.fileName(), e.Pos, e.Err)
}

// prettyPrint prints the error in a pretty way and returns true if it was printed successfully.
func (e *Error) prettyPrint(w io.Writer, opts PrintListingOptions) (handled bool, writeErr error) {
	var (
		iterErr = e.Err
		locErr  = e
	)
	for {
		e, ok := errors.Into[*Error](iterErr)
		if !ok || e.Pos.Line == 0 {
			break
		}
		locErr = e
		iterErr = e.Err
	}
	if locErr.Pos.Line != 0 {
		writeErr = e.File.PrintListing(w, locErr.Err.Error(), locErr.Pos, opts)
		return true, writeErr
	}

	return false, nil
}

func printYAMLError(w io.Writer, err error, f File, opts PrintListingOptions) (handled bool, writeErr error) {
	const printLimit = 5

	if e, ok := errors.Into[*yaml.SyntaxError](err); ok {
		loc := Position{
			Line: e.Line,
		}
		writeErr = f.PrintListing(w, e.Msg, loc, opts)
		return true, writeErr
	}

	if e, ok := errors.Into[*yaml.TypeError](err); ok {
		printed := 0
		for _, e := range multierr.Errors(e.Group) {
			if printed >= printLimit {
				break
			}
			if e, ok := errors.Into[*yaml.UnmarshalError](e); ok && e.Node != nil {
				loc := Position{
					Line:   e.Node.Line,
					Column: e.Node.Column,
					Node:   e.Node,
				}
				multierr.AppendInto(&writeErr, f.PrintListing(w, e.Err.Error(), loc, opts))
				printed++
			}
		}
		// Consider the error as handled if it is printed at least once.
		return printed > 0, writeErr
	}

	return false, nil
}

// PrintPrettyError prints the error in a pretty way and returns true if it was printed successfully.
func PrintPrettyError(w io.Writer, color bool, err error) bool {
	opts := PrintListingOptions{
		Context: 5,
	}
	if !color {
		opts = opts.WithoutColor()
	}

	v, ok := errors.Into[*Error](err)
	if !ok {
		return false
	}

	// TODO(tdakkota): handle errors?
	if handled, _ := printYAMLError(w, v.Err, v.File, opts); handled {
		return true
	}

	handled, _ := v.prettyPrint(w, opts)
	return handled
}
