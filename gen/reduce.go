package gen

import (
	"reflect"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/openapi"
)

// reduceDefault implements convenient errors, representing common default
// response as error instead of variant of each response.
func (g *Generator) reduceDefault(ops []*openapi.Operation) error {
	if len(ops) < 1 {
		return nil
	}

	// Compare first default response to others.
	first := ops[0]
	if first.Responses == nil || first.Responses["default"] == nil {
		return nil
	}
	d := first.Responses["default"]
	if d.Ref == "" {
		// Not supported.
		return nil
	}
	for _, spec := range ops[1:] {
		if !reflect.DeepEqual(spec.Responses["default"], d) {
			return nil
		}
	}

	ctx := &genctx{
		jsonptr: newJSONPointer("x-ogen-reduce-default"),
		global:  g.tstorage,
		local:   g.tstorage,
	}

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
