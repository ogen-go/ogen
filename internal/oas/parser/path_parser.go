package parser

import (
	"fmt"

	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen/internal/oas"
)

// ErrPathParameterNotSpecified indicates that the path parameter
// is not declared in the Operation parameters section.
type ErrPathParameterNotSpecified struct {
	ParamName string
}

func (e ErrPathParameterNotSpecified) Error() string {
	return fmt.Sprintf("path parameter '%s' not found in parameters", e.ParamName)
}

type pathParser struct {
	path   string           // immutable
	params []*oas.Parameter // immutable

	parts []oas.PathPart // parsed parts
	part  []rune         // current part
	param bool           // current part is param name?
}

func parsePath(path string, params []*oas.Parameter) ([]oas.PathPart, error) {
	return (&pathParser{
		path:   path,
		params: params,
	}).Parse()
}

func (p *pathParser) Parse() ([]oas.PathPart, error) {
	err := p.parse()
	return p.parts, err
}

func (p *pathParser) parse() error {
	for _, r := range p.path {
		switch r {
		case '/':
			if p.param {
				return xerrors.Errorf("invalid path: %s", p.path)
			}
			p.part = append(p.part, r)

		case '{':
			if p.param {
				return xerrors.Errorf("invalid path: %s", p.path)
			}
			if err := p.push(); err != nil {
				return err
			}
			p.param = true

		case '}':
			if !p.param {
				return xerrors.Errorf("invalid path: %s", p.path)
			}
			if err := p.push(); err != nil {
				return err
			}
			p.param = false

		default:
			p.part = append(p.part, r)
		}
	}

	if p.param {
		return xerrors.Errorf("invalid path: %s", p.path)
	}

	return p.push()
}

func (p *pathParser) push() error {
	if len(p.part) == 0 {
		return nil
	}
	defer func() { p.part = nil }()

	if !p.param {
		p.parts = append(p.parts, oas.PathPart{Raw: string(p.part)})
		return nil
	}

	param, found := p.lookupParam(string(p.part))
	if !found {
		return &ErrPathParameterNotSpecified{
			ParamName: string(p.part),
		}
	}

	p.parts = append(p.parts, oas.PathPart{Param: param})
	return nil
}

func (p *pathParser) lookupParam(name string) (*oas.Parameter, bool) {
	for _, p := range p.params {
		if p.Name == name && p.In == oas.LocationPath {
			return p, true
		}
	}
	return nil, false
}
