package parser

import (
	"fmt"

	"github.com/go-faster/errors"
	"golang.org/x/exp/slices"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/jsonpointer"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseServers(servers []ogen.Server, ctx *jsonpointer.ResolveCtx) ([]openapi.Server, error) {
	if len(servers) == 0 {
		return nil, nil
	}
	r := make([]openapi.Server, len(servers))
	for i, s := range servers {
		srv, err := p.parseServer(s, ctx)
		if err != nil {
			return nil, err
		}
		r[i] = srv
	}
	return r, nil
}

func (p *parser) parseServer(s ogen.Server, ctx *jsonpointer.ResolveCtx) (_ openapi.Server, rerr error) {
	locator := s.Common.Locator
	defer func() {
		rerr = p.wrapLocation(ctx.LastLoc(), locator, rerr)
	}()

	// Validate variables.
	for _, v := range s.Variables {
		locator := v.Common.Locator
		if v.Default == "" {
			locator := locator.Field("default")
			err := errors.New("default MUST be set and not be empty")
			return openapi.Server{}, p.wrapLocation(ctx.LastLoc(), locator, err)
		}

		if len(v.Enum) > 0 {
			locator := locator.Field("enum")
			if !slices.Contains(v.Enum, v.Default) {
				err := errors.Errorf("the default value %q MUST exist in the enum's values", v.Default)
				return openapi.Server{}, p.wrapLocation(ctx.LastLoc(), locator, err)
			}
			for i, e := range v.Enum {
				locator := locator.Index(i)
				found := slices.Index(v.Enum, e)
				if found < 0 {
					panic(fmt.Sprintf("unreachable: slice %#v doesn't contain %q", v.Enum, e))
				}
				if found != i {
					err := errors.Errorf("enum MUST NOT contain duplicate values: %q", e)
					return openapi.Server{}, p.wrapLocation(ctx.LastLoc(), locator, err)
				}
			}
		}
	}

	// Validate URL.
	u, err := func() (openapi.Path, error) {
		if s.URL == "" {
			return nil, errors.New("server URL must not be empty")
		}
		return parseServerURL(s.URL, func(name string) (*openapi.Parameter, bool) {
			v, ok := s.Variables[name]
			if !ok {
				return nil, false
			}

			schema := &jsonschema.Schema{
				Type:       jsonschema.String,
				Default:    v.Default,
				DefaultSet: true,
			}
			if len(v.Enum) > 0 {
				schema.Enum = make([]any, len(v.Enum))
				for i, e := range v.Enum {
					schema.Enum[i] = e
				}
			}
			return &openapi.Parameter{
				Name:     name,
				Schema:   schema,
				In:       openapi.LocationPath,
				Style:    openapi.PathStyleSimple,
				Required: true,
				Locator:  v.Common.Locator,
			}, true
		})
	}()
	if err != nil {
		locator := locator.Field("url")
		return openapi.Server{}, p.wrapLocation(ctx.LastLoc(), locator, err)
	}

	return openapi.Server{
		Template:    u,
		Description: s.Description,
	}, nil
}
