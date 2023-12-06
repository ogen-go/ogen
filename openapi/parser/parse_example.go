package parser

import (
	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonpointer"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseExample(e *ogen.Example, ctx *jsonpointer.ResolveCtx) (_ *openapi.Example, rerr error) {
	if e == nil {
		return nil, nil
	}
	locator := e.Common.Locator
	defer func() {
		rerr = p.wrapLocation(p.file(ctx), locator, rerr)
	}()
	if ref := e.Ref; ref != "" {
		resolved, err := p.resolveExample(ref, ctx)
		if err != nil {
			return nil, p.wrapRef(p.file(ctx), locator, err)
		}
		return resolved, nil
	}

	return &openapi.Example{
		Summary:       e.Summary,
		Description:   e.Description,
		Value:         e.Value,
		ExternalValue: e.ExternalValue,
		Pointer:       locator.Pointer(p.file(ctx)),
	}, nil
}
