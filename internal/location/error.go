package location

import (
	"fmt"
	"io"
	"strings"

	"github.com/go-faster/errors"
	"github.com/go-faster/yaml"
	"go.uber.org/multierr"
)

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

// FormatError implements errors.Formatter.
func (e *Error) FormatError(p errors.Printer) error {
	p.Printf("at %s", e.Pos.WithFilename(e.File.humanName()))
	return e.Err
}

// Format implements fmt.Formatter.
func (e *Error) Format(s fmt.State, verb rune) {
	errors.FormatError(e, s, verb)
}

// Error implements error.
func (e *Error) Error() string {
	return fmt.Sprintf("at %s: %s", e.Pos.WithFilename(e.File.humanName()), e.Err)
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

// Report is element of MultiError container.
type Report struct {
	File File
	Pos  Position
	Msg  string
}

// String returns textual represntation of Report.
func (r Report) String() string {
	return fmt.Sprintf("at %s: %s", r.Pos.WithFilename(r.File.humanName()), r.Msg)
}

var _ interface {
	errors.Formatter
	fmt.Formatter
	error
} = (*MultiError)(nil)

// MultiError contains multiple Reports.
type MultiError struct {
	Reports []Report
}

func (e *MultiError) printSingle(printf func(format string, args ...any)) {
	switch len(e.Reports) {
	case 0:
		printf("empty error")
	case 1:
		printf("%s", e.Reports[0].String())
	default:
		for _, r := range e.Reports {
			printf("- at %s\n", r.String())
		}
	}
}

// FormatError implements errors.Formatter.
func (e *MultiError) FormatError(p errors.Printer) error {
	e.printSingle(p.Printf)
	return nil
}

// Format implements fmt.Formatter.
func (e *MultiError) Format(s fmt.State, verb rune) {
	errors.FormatError(e, s, verb)
}

// Error implements error.
func (e *MultiError) Error() string {
	var sb strings.Builder
	e.printSingle(func(format string, args ...any) {
		fmt.Fprintf(&sb, format, args...)
	})
	return sb.String()
}

const printLimit = 5

// prettyPrint prints the error in a pretty way and returns true if it was printed successfully.
func (e *MultiError) prettyPrint(w io.Writer, opts PrintListingOptions) (handled bool, writeErr error) {
	printed := 0
	for _, r := range e.Reports {
		if printed >= printLimit {
			break
		}

		// TODO(tdakkota): print close location together
		f := r.File
		multierr.AppendInto(&writeErr, f.PrintListing(w, r.Msg, r.Pos, opts))
		printed++
	}

	return printed > 0, writeErr
}

func printYAMLError(w io.Writer, err error, f File, opts PrintListingOptions) (handled bool, writeErr error) {
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

	// TODO(tdakkota): handle write errors?
	me, ok := errors.Into[*MultiError](err)
	if ok {
		if handled, _ := me.prettyPrint(w, opts); handled {
			return true
		}
	}

	e, ok := errors.Into[*Error](err)
	if !ok {
		return false
	}

	if handled, _ := printYAMLError(w, e.Err, e.File, opts); handled {
		return true
	}

	handled, _ := e.prettyPrint(w, opts)
	return handled
}
