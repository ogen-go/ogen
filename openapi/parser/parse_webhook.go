package parser

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/xmaps"
	"github.com/ogen-go/ogen/jsonpointer"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseWebhook(name string, item *ogen.PathItem, ctx *jsonpointer.ResolveCtx) (openapi.Webhook, error) {
	// FIXME(tdakkota): we are passing "/" path, but webhook has no path.
	pi, err := p.parsePathItem(unparsedPath{path: "/"}, item, ctx)
	if err != nil {
		return openapi.Webhook{}, errors.Wrap(err, "parse pathItem")
	}
	return openapi.Webhook{
		Name:       name,
		Operations: pi,
		Pointer:    item.Common.Locator.Pointer(p.file(ctx)),
	}, nil
}

func (p *parser) parseWebhooks(webhooks map[string]*ogen.PathItem) (r []openapi.Webhook, rerr error) {
	if len(webhooks) == 0 {
		return nil, nil
	}
	var (
		locator = p.rootLoc.Field("webhooks")
		ctx     = p.resolveCtx()
	)
	defer func() {
		rerr = p.wrapLocation(p.file(ctx), locator, rerr)
	}()
	if err := p.requireMinorVersion("webhooks", 1); err != nil {
		return nil, err
	}

	r = make([]openapi.Webhook, 0, len(webhooks))
	for _, name := range xmaps.SortedKeys(webhooks) {
		item := webhooks[name]
		webhook, err := p.parseWebhook(name, item, ctx)
		if err != nil {
			err := errors.Wrapf(err, "parse webhook %q", name)
			return nil, p.wrapLocation(p.file(ctx), locator.Field(name), err)
		}
		r = append(r, webhook)
	}
	return r, nil
}
