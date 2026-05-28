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

func (g *genctx) saveWType(parent, ref jsonschema.Ref, headers map[string]*ir.Parameter, t *ir.Type) error {
	return g.local.saveWType(parent, ref, headers, t)
}

func (g *genctx) saveParameter(ref jsonschema.Ref, r *ir.Parameter) error {
	return g.local.saveParameter(ref, r)
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

func (g *genctx) lookupWType(parent, ref jsonschema.Ref, headers map[string]*ir.Parameter) (*ir.Type, bool) {
	key := [3]jsonschema.Ref{parent, ref, headersRef(headers)}
	if t, ok := g.global.wtypes[key]; ok {
		return t, true
	}
	if t, ok := g.local.wtypes[key]; ok {
		return t, true
	}
	return nil, false
}

func (g *genctx) lookupType(name string) (*ir.Type, bool) {
	if t, ok := g.global.types[name]; ok {
		return t, true
	}
	if t, ok := g.local.types[name]; ok {
		return t, true
	}
	return nil, false
}

func (g *genctx) lookupParameter(ref jsonschema.Ref) (*ir.Parameter, bool) {
	if r, ok := g.global.parameters[ref]; ok {
		return r, true
	}
	if r, ok := g.local.parameters[ref]; ok {
		return r, true
	}
	return nil, false
}
