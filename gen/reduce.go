package gen

import (
	"reflect"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/ogen-go/ogen/internal/location"
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
		return reduceFailed(`operation has no "default" response`, first)
	}
	// TODO(tdakkota): handle cases when response share same JSON Schema and all
	// 	other fields are the same.
	if d.Ref.IsZero() {
		// Not supported.
		return reduceFailed(`response must be a reference`, d)
	}
	for _, op := range ops[1:] {
		switch other := op.Responses.Default; {
		case other == nil:
			return reduceFailed(`operation has no "default" response`, op)
		case !reflect.DeepEqual(other, d):
			return reduceFailed(`"default" response is different`, other)
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
