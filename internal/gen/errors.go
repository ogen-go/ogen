package gen

import (
	"fmt"

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

func (e ErrNotImplemented) Error() string {
	return e.Name + " not implemented"
}

func (g *Generator) checkErr(err error) error {
	{
		var notImplementedErr *ErrNotImplemented
		if xerrors.As(err, &notImplementedErr) {
			if g.opt.IgnoreNotImplemented {
				return nil
			}
		}
	}
	{
		var paramErr *ErrPathParameterNotSpecified
		if xerrors.As(err, &paramErr) {
			if g.opt.IgnoreUnspecifiedParams {
				return nil
			}
		}
	}

	return err
}
