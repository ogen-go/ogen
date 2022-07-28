package parser

import (
	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseExample(e *ogen.Example, ctx *resolveCtx) (_ *openapi.Example, rerr error) {
	if e == nil {
		return nil, nil
	}
	defer func() {
		rerr = p.wrapLocation(ctx.lastLoc(), e.Locator, rerr)
	}()
	if ref := e.Ref; ref != "" {
		resolved, err := p.resolveExample(ref, ctx)
		if err != nil {
			return nil, p.wrapRef(ctx.lastLoc(), e.Locator, err)
		}
		return resolved, nil
	}

	return &openapi.Example{
		Summary:       e.Summary,
		Description:   e.Description,
		Value:         e.Value,
		ExternalValue: e.ExternalValue,
		Locator:       e.Locator,
	}, nil
}
