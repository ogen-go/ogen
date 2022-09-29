package parser

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/jsonpointer"
	"github.com/ogen-go/ogen/internal/xmaps"
	"github.com/ogen-go/ogen/openapi"
)

func (p *parser) parseWebhook(name string, item *ogen.PathItem, ctx *jsonpointer.ResolveCtx) (openapi.Webhook, error) {
	operations := make(map[string]struct{})
	// FIXME(tdakkota): we are passing "/" path, but webhook has no path.
	pi, err := p.parsePathItem("/", item, operations, ctx)
	if err != nil {
		return openapi.Webhook{}, errors.Wrap(err, "parse pathItem")
	}
	return openapi.Webhook{
		Name:       name,
		Operations: pi,
	}, nil
}

func (p *parser) parseWebhooks(webhooks map[string]*ogen.PathItem) (r []openapi.Webhook, rerr error) {
	var (
		locator = p.rootLoc.Field("webhooks")
		ctx     = jsonpointer.NewResolveCtx(p.depthLimit)
	)
	defer func() {
		rerr = p.wrapLocation(ctx.LastLoc(), locator, rerr)
	}()

	for _, name := range xmaps.SortedKeys(webhooks) {
		item := webhooks[name]
		webhook, err := p.parseWebhook(name, item, ctx)
		if err != nil {
			err := errors.Wrapf(err, "parse webhook %q", name)
			return nil, p.wrapLocation(ctx.LastLoc(), locator.Field(name), err)
		}
		r = append(r, webhook)
	}
	return r, nil
}
