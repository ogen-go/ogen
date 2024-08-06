package gen

import (
	"fmt"
	"slices"
	"strings"

	"github.com/go-faster/errors"
	"github.com/ogen-go/ogen/gen/ir"
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

// ErrFieldsDiscriminatorInference reports fields discriminator inference failure.
type ErrFieldsDiscriminatorInference struct {
	Sum   *ir.Type
	Types []BadVariant
}

func (e *ErrFieldsDiscriminatorInference) unimplemented() {}

// Error implements error.
func (e *ErrFieldsDiscriminatorInference) Error() string {
	names := make([]string, len(e.Types))
	for i, typ := range e.Types {
		names[i] = typ.Type.Name
	}
	return fmt.Sprintf("can't infer fields discriminator: [%s]", strings.Join(names, ", "))
}

// BadVariant describes a sum type variant for what we unable to infer discriminator.
type BadVariant struct {
	Type   *ir.Type
	Fields map[string][]*ir.Type
}

func (g *Generator) trySkip(err error, msg string, l position) error {
	if err == nil {
		return nil
	}
	if err := g.fail(err); err != nil {
		return err
	}

	reason := err.Error()
	if uErr, ok := errors.Into[unimplementedError](err); ok {
		reason = uErr.Error()
	}
	g.log.WithOptions(zap.AddCallerSkip(1)).Info(msg,
		zapPosition(l),
		zap.String("reason_error", reason),
	)
	return nil
}

func (g *Generator) fail(err error) error {
	if err == nil {
		return nil
	}
	hasAll := slices.Contains(g.opt.IgnoreNotImplemented, "all")
	handle := func(name string, err error) error {
		if hook := g.opt.NotImplementedHook; hook != nil {
			hook(name, err)
		}
		if hasAll || slices.Contains(g.opt.IgnoreNotImplemented, name) {
			return nil
		}
		return err
	}
	if notImplementedErr, ok := errors.Into[*ErrNotImplemented](err); ok {
		return handle(notImplementedErr.Name, err)
	}
	if _, ok := errors.Into[*ErrUnsupportedContentTypes](err); ok {
		const name = "unsupported content types"
		return handle(name, err)
	}
	if _, ok := errors.Into[*ErrFieldsDiscriminatorInference](err); ok {
		const name = "discriminator inference"
		return handle(name, err)
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
