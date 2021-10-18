package gen

import (
	"fmt"
	"strings"

	"golang.org/x/xerrors"
)

type ErrPathParameterNotSpecified struct {
	ParamName string
}

func (e ErrPathParameterNotSpecified) Error() string {
	return fmt.Sprintf("path parameter '%s' not found in parameters", e.ParamName)
}

type ErrNotImplemented struct {
	Name string
}

func (e *ErrNotImplemented) Error() string {
	return e.Name + " not implemented"
}

func (g *Generator) checkErr(err error) error {
	var notImplementedErr *ErrNotImplemented
	if xerrors.As(err, &notImplementedErr) {
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
	var paramErr *ErrPathParameterNotSpecified
	if xerrors.As(err, &paramErr) {
		if g.opt.IgnoreUnspecifiedParams {
			return nil
		}
	}

	return err
}
