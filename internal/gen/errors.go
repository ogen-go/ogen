package gen

import (
	"fmt"

	"golang.org/x/xerrors"
)

var (
	ErrEnumsNotImplemented  = xerrors.New("enums not implemented")
	ErrOneOfNotImplemented  = xerrors.New("oneOf not implemented")
	ErrAnyOfNotImplemented  = xerrors.New("anyOf not implemented")
	ErrAllOfNotImplemented  = xerrors.New("allOf not implemented")
	ErrUnsupportedParameter = xerrors.New("parameter type not supported")
)

type PathParameterNotSpecified struct {
	ParamName string
}

func (p PathParameterNotSpecified) Error() string {
	return fmt.Sprintf("path parameter '%s' not found in parameters", p.ParamName)
}

func (g *Generator) checkErr(err error) error {
	if xerrors.Is(err, ErrEnumsNotImplemented) {
		if g.opt.IgnoreEnums {
			return nil
		}
	}
	if xerrors.Is(err, ErrOneOfNotImplemented) {
		if g.opt.IgnoreOneOf {
			return nil
		}
	}
	if xerrors.Is(err, ErrAnyOfNotImplemented) {
		if g.opt.IgnoreAnyOf {
			return nil
		}
	}
	if xerrors.Is(err, ErrAllOfNotImplemented) {
		if g.opt.IgnoreAllOf {
			return nil
		}
	}
	if xerrors.Is(err, ErrUnsupportedParameter) {
		if g.opt.IgnoreUnsupportedParams {
			return nil
		}
	}
	{
		var paramErr *PathParameterNotSpecified
		if xerrors.As(err, &paramErr) {
			if g.opt.IgnoreUnspecifiedParams {
				return nil
			}
		}
	}

	return err
}
