package gen

import (
	"fmt"
	"strings"

	"github.com/go-faster/errors"
)

type ErrNotImplemented struct {
	Name string
}

func (e *ErrNotImplemented) Error() string {
	return e.Name + " not implemented"
}

type ErrUnsupportedContentTypes struct {
	ContentTypes []string
}

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

type ErrParseSpec struct {
	err error
}

func (e *ErrParseSpec) Unwrap() error {
	return e.err
}

func (e *ErrParseSpec) Error() string {
	return fmt.Sprintf("parse spec: %s", e.err)
}

type ErrBuildRouter struct {
	err error
}

func (e *ErrBuildRouter) Unwrap() error {
	return e.err
}

func (e *ErrBuildRouter) Error() string {
	return fmt.Sprintf("build router: %s", e.err)
}
