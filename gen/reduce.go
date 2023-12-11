package gen

import (
	"encoding/json"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/ogen-go/ogen/gen/ir"
	ogenjson "github.com/ogen-go/ogen/json"
	"github.com/ogen-go/ogen/location"
	"github.com/ogen-go/ogen/openapi"
)

// reduceDefault implements convenient errors, representing common default
// response as error instead of variant of each response.
func (g *Generator) reduceDefault(ops []*openapi.Operation) error {
	log := g.log.Named("convenient")
	if g.opt.ConvenientErrors.IsDisabled() {
		log.Info("Convenient errors are disabled, skip reduce")
		return nil
	}
	reduceFailed := func(msg string, p position) error {
		if g.opt.ConvenientErrors.IsForced() {
			err := errors.Wrap(errors.New(msg), "can't reduce to convenient error")

			pos, ok := p.Position()
			if !ok {
				return err
			}

			return &location.Error{
				File: p.File(),
				Pos:  pos,
				Err:  err,
			}
		}
		log.Info("Convenient errors are not available",
			zap.String("reason", msg),
			zapPosition(p),
		)
		return nil
	}

	if len(ops) < 1 {
		return nil
	}

	// Compare first default response to others.
	//
	// TODO(tdakkota): reduce by 4XX/5XX?
	first := ops[0]
	d := first.Responses.Default
	if d == nil {
		return reduceFailed(`operation has no "default" response`, first.Responses)
	}
	switch {
	case len(d.Content) < 1:
		// TODO(tdakkota): point to "content", not to the entire response
		return reduceFailed(`response is no-content`, d)
	case len(d.Content) > 1:
		// TODO(tdakkota): point to "content", not to the entire response
		return reduceFailed(`response is multi-content`, d)
	}
	{
		var ct ir.Encoding
		for key := range d.Content {
			ct = ir.Encoding(key)
			break
		}
		if override, ok := g.opt.ContentTypeAliases[string(ct)]; ok {
			ct = override
		}
		if !ct.JSON() {
			return reduceFailed(`response content must be JSON`, d)
		}
	}

	compareResponses := func(a, b *openapi.Response) bool {
		// Compile time check to not forget to update compareResponses.
		type check struct {
			Ref         openapi.Ref
			Description string
			Headers     map[string]*openapi.Header
			Content     map[string]*openapi.MediaType

			location.Pointer `json:"-" yaml:"-"`
		}
		var (
			_ = (*check)(a)
			_ = (*check)(b)
		)

		switch {
		case a == b:
			return true
		case a == nil || b == nil:
			return false
		}

		// Set of fields to compare.
		type compare struct {
			Ref     openapi.Ref
			Headers map[string]*openapi.Header
			Content map[string]*openapi.MediaType
		}

		x, err := json.Marshal(compare{
			a.Ref,
			a.Headers,
			a.Content,
		})
		if err != nil {
			return false
		}

		y, err := json.Marshal(compare{
			b.Ref,
			b.Headers,
			b.Content,
		})
		if err != nil {
			return false
		}

		equal, _ := ogenjson.Equal(x, y)
		return equal
	}

	for _, op := range ops[1:] {
		switch other := op.Responses.Default; {
		case other == nil:
			return reduceFailed(`operation has no "default" response`, op.Responses)
		case !compareResponses(d, other):
			return reduceFailed(`response is different`, other)
		}
	}

	ctx := &genctx{
		global: g.tstorage,
		local:  g.tstorage,
	}

	log.Info("Generating convenient error response", zapPosition(d))
	resp, err := g.responseToIR(ctx, "ErrResp", "reduced default response", d, true)
	if err != nil {
		return errors.Wrap(err, "default")
	}

	hasJSON := false
	for _, media := range resp.Contents {
		if media.Encoding.JSON() {
			hasJSON = true
			break
		}
	}
	if resp.NoContent != nil || len(resp.Contents) > 1 || !hasJSON {
		return errors.Wrap(err, "too complicated to reduce default error")
	}

	g.errType = resp
	return nil
}
