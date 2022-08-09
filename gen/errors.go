package gen

import (
	"fmt"
	"strings"

	"github.com/go-faster/errors"
	"go.uber.org/zap"
)

type unimplementedError interface {
	unimplemented()
	error
}

var (
	_ = []interface {
		errors.Wrapper
		errors.Formatter
		fmt.Formatter
		error
	}{
		(*ErrParseSpec)(nil),
		(*ErrBuildRouter)(nil),
		(*ErrGoFormat)(nil),
	}
	_ = []interface {
		error
		unimplementedError
	}{
		(*ErrNotImplemented)(nil),
		(*ErrUnsupportedContentTypes)(nil),
	}
)

// ErrNotImplemented reports that feature is not implemented.
type ErrNotImplemented struct {
	Name string
}

func (e *ErrNotImplemented) unimplemented() {}

// Error implements error.
func (e *ErrNotImplemented) Error() string {
	return e.Name + " not implemented"
}

// ErrUnsupportedContentTypes reports that ogen does not support such content-type(s).
type ErrUnsupportedContentTypes struct {
	ContentTypes []string
}

func (e *ErrUnsupportedContentTypes) unimplemented() {}

// Error implements error.
func (e *ErrUnsupportedContentTypes) Error() string {
	return fmt.Sprintf("unsupported content types: [%s]", strings.Join(e.ContentTypes, ", "))
}

func (g *Generator) trySkip(err error, msg string, l locatable) error {
	if err == nil {
		return nil
	}
	if err := g.fail(err); err != nil {
		return err
	}

	reason := err.Error()
	if uErr := unimplementedError(nil); errors.As(err, &uErr) {
		reason = uErr.Error()
	}
	g.log.Info(msg,
		g.zapLocation(l),
		zap.String("reason_error", reason),
	)
	return nil
}

func (g *Generator) fail(err error) error {
	if err == nil {
		return nil
	}

	hook := g.opt.NotImplementedHook
	if hook == nil {
		hook = func(string, error) {}
	}

	var notImplementedErr *ErrNotImplemented
	if errors.As(err, &notImplementedErr) {
		hook(notImplementedErr.Name, err)
		for _, s := range g.opt.IgnoreNotImplemented {
			s = strings.TrimSpace(s)
			if s == "all" {
				return nil
			}
			if s == notImplementedErr.Name {
				return nil
			}
		}
	}

	var ctypesErr *ErrUnsupportedContentTypes
	if errors.As(err, &ctypesErr) {
		hook("unsupported content types", err)
		for _, s := range g.opt.IgnoreNotImplemented {
			s = strings.TrimSpace(s)
			if s == "all" || s == "unsupported content types" {
				return nil
			}
		}
	}
	return err
}

// ErrParseSpec reports that specification parsing failed.
type ErrParseSpec struct {
	err error
}

// Unwrap implements errors.Wrapper.
func (e *ErrParseSpec) Unwrap() error {
	return e.err
}

// FormatError implements errors.Formatter.
func (e *ErrParseSpec) FormatError(p errors.Printer) (next error) {
	p.Print("parse spec")
	return e.err
}

// Format implements fmt.Formatter.
func (e *ErrParseSpec) Format(s fmt.State, verb rune) {
	errors.FormatError(e, s, verb)
}

// Error implements error.
func (e *ErrParseSpec) Error() string {
	return fmt.Sprintf("parse spec: %s", e.err)
}

// ErrBuildRouter reports that route tree building failed.
type ErrBuildRouter struct {
	err error
}

// Unwrap implements errors.Wrapper.
func (e *ErrBuildRouter) Unwrap() error {
	return e.err
}

// FormatError implements errors.Formatter.
func (e *ErrBuildRouter) FormatError(p errors.Printer) (next error) {
	p.Print("build router")
	return e.err
}

// Format implements fmt.Formatter.
func (e *ErrBuildRouter) Format(s fmt.State, verb rune) {
	errors.FormatError(e, s, verb)
}

// Error implements error.
func (e *ErrBuildRouter) Error() string {
	return fmt.Sprintf("build router: %s", e.err)
}

// ErrGoFormat reports that generated code formatting failed.
type ErrGoFormat struct {
	err error
}

// Unwrap implements errors.Wrapper.
func (e *ErrGoFormat) Unwrap() error {
	return e.err
}

// FormatError implements errors.Formatter.
func (e *ErrGoFormat) FormatError(p errors.Printer) (next error) {
	p.Print("goimports")
	return e.err
}

// Format implements fmt.Formatter.
func (e *ErrGoFormat) Format(s fmt.State, verb rune) {
	errors.FormatError(e, s, verb)
}

// Error implements error.
func (e *ErrGoFormat) Error() string {
	return fmt.Sprintf("goimports: %s", e.err)
}
