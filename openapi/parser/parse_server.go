package parser

import (
	"fmt"
	"go/token"
	"slices"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonpointer"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseServers(servers []ogen.Server, ctx *jsonpointer.ResolveCtx) ([]openapi.Server, error) {
	if len(servers) == 0 {
		return nil, nil
	}
	var (
		r     = make([]openapi.Server, len(servers))
		dedup = map[string]struct{}{}
	)
	for i, s := range servers {
		srv, err := p.parseServer(s, dedup, ctx)
		if err != nil {
			return nil, err
		}
		r[i] = srv
	}
	return r, nil
}

func (p *parser) parseServer(
	s ogen.Server,
	dedup map[string]struct{},
	ctx *jsonpointer.ResolveCtx,
) (_ openapi.Server, rerr error) {
	locator := s.Common.Locator
	defer func() {
		rerr = p.wrapLocation(p.file(ctx), locator, rerr)
	}()

	// Validate variables.
	for _, v := range s.Variables {
		locator := v.Common.Locator
		if len(v.Enum) > 0 {
			locator := locator.Field("enum")
			if !slices.Contains(v.Enum, v.Default) {
				err := errors.Errorf("the default value %q MUST exist in the enum's values", v.Default)
				return openapi.Server{}, p.wrapLocation(p.file(ctx), locator, err)
			}
			for i, e := range v.Enum {
				locator := locator.Index(i)
				found := slices.Index(v.Enum, e)
				if found < 0 {
					panic(fmt.Sprintf("unreachable: slice %#v doesn't contain %q", v.Enum, e))
				}
				if found != i {
					err := errors.Errorf("enum MUST NOT contain duplicate values: %q", e)
					return openapi.Server{}, p.wrapLocation(p.file(ctx), locator, err)
				}
			}
		}
	}

	// Validate URL.
	u, err := func() (openapi.ServerURL, error) {
		if s.URL == "" {
			return nil, errors.New("server URL must not be empty")
		}
		return parseServerURL(s.URL, func(name string) (sv openapi.ServerVariable, _ bool) {
			v, ok := s.Variables[name]
			if !ok {
				return sv, false
			}

			return openapi.ServerVariable{
				Name:        name,
				Description: v.Description,
				Default:     v.Default,
				Enum:        slices.Clone(v.Enum),
			}, true
		})
	}()
	if err != nil {
		locator := locator.Field("url")
		return openapi.Server{}, p.wrapLocation(p.file(ctx), locator, err)
	}

	server := openapi.Server{
		Template:    u,
		Description: s.Description,
	}

	// Parse extensions.
	//
	// TODO(tdakkota): describe extensions somewhere, it would be nice to have machine-readable
	// 	description of extensions, their types, and their validation rules.
	const nameKey = "x-ogen-server-name"
	if nameNode, ok := s.Common.Extensions[nameKey]; ok {
		if err := func() error {
			if err := nameNode.Decode(&server.Name); err != nil {
				return err
			}

			name := server.Name
			switch {
			case name == "":
				return errors.New("server name must not be empty")
			case !token.IsIdentifier(name + "Server"):
				// Ensure that ${name}Server is a valid Go identifier.
				return errors.Errorf("server name %q cannot be used as a Go identifier", name)
			}
			if _, ok := dedup[name]; ok {
				return errors.Errorf("server name %q is not unique", name)
			}
			dedup[name] = struct{}{}

			return nil
		}(); err != nil {
			locator := locator.Field(nameKey)
			return openapi.Server{}, p.wrapLocation(p.file(ctx), locator, err)
		}
	}

	return server, nil
}
