package gen

import (
	"fmt"
	"strings"

	"github.com/go-faster/errors"
)

// ErrNotImplemented reports that feature is not implemented.
type ErrNotImplemented struct {
	Name string
}

// Error implements error.
func (e *ErrNotImplemented) Error() string {
	return e.Name + " not implemented"
}

// ErrUnsupportedContentTypes reports that ogen does not support such content-type(s).
type ErrUnsupportedContentTypes struct {
	ContentTypes []string
}

// Error implements error.
func (e *ErrUnsupportedContentTypes) Error() string {
	return fmt.Sprintf("unsupported content types: [%s]", strings.Join(e.ContentTypes, ", "))
}

func (g *Generator) fail(err error) error {
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

// Error implements error.
func (e *ErrBuildRouter) Error() string {
	return fmt.Sprintf("build router: %s", e.err)
}
