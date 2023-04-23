package gen

import (
	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/jsonschema"
)

// genctx is a generation context.
type genctx struct {
	global *tstorage // readonly
	local  *tstorage
}

func (g *genctx) saveType(t *ir.Type) error {
	return g.local.saveType(t)
}

func (g *genctx) saveRef(ref jsonschema.Ref, e ir.Encoding, t *ir.Type) error {
	return g.local.saveRef(ref, e, t)
}

func (g *genctx) lookupRef(ref jsonschema.Ref, e ir.Encoding) (*ir.Type, bool) {
	key := schemaKey{ref, e}
	if t, ok := g.global.refs[key]; ok {
		return t, true
	}
	if t, ok := g.local.refs[key]; ok {
		return t, true
	}
	return nil, false
}

func (g *genctx) saveResponse(ref jsonschema.Ref, r *ir.Response) error {
	return g.local.saveResponse(ref, r)
}

func (g *genctx) saveWType(ref jsonschema.Ref, t *ir.Type) error {
	return g.local.saveWType(ref, t)
}

func (g *genctx) lookupResponse(ref jsonschema.Ref) (*ir.Response, bool) {
	if r, ok := g.global.responses[ref]; ok {
		return r, true
	}
	if r, ok := g.local.responses[ref]; ok {
		return r, true
	}
	return nil, false
}

func (g *genctx) lookupWType(ref jsonschema.Ref) (*ir.Type, bool) {
	if t, ok := g.global.wtypes[ref]; ok {
		return t, true
	}
	if t, ok := g.local.wtypes[ref]; ok {
		return t, true
	}
	return nil, false
}
