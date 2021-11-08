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

func (g *Generator) shouldFail(err error) bool {
	var notImplementedErr *ErrNotImplemented
	if errors.As(err, &notImplementedErr) {
		for _, s := range g.opt.IgnoreNotImplemented {
			s = strings.TrimSpace(s)
			if s == "all" {
				return false
			}
			if s == notImplementedErr.Name {
				return false
			}
		}
	}

	var ctypesErr *ErrUnsupportedContentTypes
	if errors.As(err, &ctypesErr) {
		for _, s := range g.opt.IgnoreNotImplemented {
			s = strings.TrimSpace(s)
			if s == "all" || s == "unsupported content types" {
				return false
			}
		}
	}
	return true
}
