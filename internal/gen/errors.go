package gen

import (
	"strings"

	"github.com/ogen-go/errors"
)

type ErrNotImplemented struct {
	Name string
}

func (e *ErrNotImplemented) Error() string {
	return e.Name + " not implemented"
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
	return true
}
