package gen

import (
	"strings"

	"golang.org/x/xerrors"
)

type ErrNotImplemented struct {
	Name string
}

func (e *ErrNotImplemented) Error() string {
	return e.Name + " not implemented"
}

func (g *Generator) shouldFail(err error) bool {
	var notImplementedErr *ErrNotImplemented
	if xerrors.As(err, &notImplementedErr) {
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
