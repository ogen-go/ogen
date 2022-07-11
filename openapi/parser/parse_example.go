package parser

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseExample(e *ogen.Example, ctx *resolveCtx) (_ *openapi.Example, rerr error) {
	if e == nil {
		return nil, nil
	}

	if ref := e.Ref; ref != "" {
		ex, err := p.resolveExample(ref, ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "resolve %q", ref)
		}
		return ex, nil
	}
	defer func() {
		rerr = p.wrapLocation(ctx.lastLoc(), &e.Locator, rerr)
	}()

	return &openapi.Example{
		Summary:       e.Summary,
		Description:   e.Description,
		Value:         e.Value,
		ExternalValue: e.ExternalValue,
		Locator:       e.Locator,
	}, nil
}
