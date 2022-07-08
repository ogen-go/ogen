package parser

import (
	"net/url"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/openapi"
)

type pathParser struct {
	path   string // immutable
	lookup func(name string) (*openapi.Parameter, bool)

	parts []openapi.PathPart // parsed parts
	part  []rune             // current part
	param bool               // current part is param name?
}

func pathID(path string) (string, error) {
	p, err := (&pathParser{
		path: path,
		lookup: func(name string) (*openapi.Parameter, bool) {
			return nil, true
		},
	}).Parse()
	if err != nil {
		return "", err
	}
	return p.ID(), nil
}

func parsePath(path string, params []*openapi.Parameter) (openapi.Path, error) {
	return (&pathParser{
		path: path,
		lookup: func(name string) (*openapi.Parameter, bool) {
			for _, p := range params {
				if p.Name == name && p.In == openapi.LocationPath {
					return p, true
				}
			}
			return nil, false
		},
	}).Parse()
}

func (p *pathParser) Parse() (openapi.Path, error) {
	err := p.parse()
	return p.parts, err
}

func (p *pathParser) parse() error {
	// Validate and unescape path.
	//
	// FIXME(tdakkota): OpenAPI spec, as always, is not clear about path validation.
	//  All we know is that it MUST start with a slash.
	// 	At the same time, https://swagger.io/docs/specification/paths-and-operations/ says that
	// 	paths must not include query parameters.
	//  In summary, we do not pass URL scheme, user info, host or query string.
	//
	u, err := url.Parse(p.path)
	if err != nil {
		return err
	}
	switch {
	case u.IsAbs() || u.Host != "" || u.User != nil:
		return errors.New("path MUST be relative")
	case !strings.HasPrefix(u.Path, "/"):
		return errors.New("path MUST begin with a forward slash")
	case u.RawQuery != "":
		return errors.New("path MUST NOT contain a query string")
	}

	for _, r := range u.Path {
		switch r {
		case '/':
			if p.param {
				return errors.Errorf("invalid path: %s", p.path)
			}
			p.part = append(p.part, r)

		case '{':
			if p.param {
				return errors.Errorf("invalid path: %s", p.path)
			}
			if err := p.push(); err != nil {
				return err
			}
			p.param = true

		case '}':
			if !p.param {
				return errors.Errorf("invalid path: %s", p.path)
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
		return errors.Errorf("invalid path: %s", p.path)
	}

	return p.push()
}

func (p *pathParser) push() error {
	if len(p.part) == 0 {
		return nil
	}
	defer func() { p.part = nil }()

	if !p.param {
		p.parts = append(p.parts, openapi.PathPart{Raw: string(p.part)})
		return nil
	}

	param, found := p.lookup(string(p.part))
	if !found {
		return errors.Errorf("path parameter not specified: %q", string(p.part))
	}

	p.parts = append(p.parts, openapi.PathPart{Param: param})
	return nil
}
