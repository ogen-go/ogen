package gen

import "github.com/ogen-go/ogen/gen/ir"

// genctx is a generation context.
type genctx struct {
	path []string

	global *tstorage // readonly
	local  *tstorage
}

func (g *genctx) appendPath(v ...string) *genctx {
	plen, pcap := len(g.path), len(g.path)
	newPath := append(g.path[:plen:pcap], v...)
	return &genctx{
		path:   newPath,
		global: g.global,
		local:  g.local,
	}
}

func (g *genctx) saveType(t *ir.Type) error {
	return g.local.saveType(t)
}

func (g *genctx) saveRef(ref string, t *ir.Type) error {
	return g.local.saveRef(ref, t)
}

func (g *genctx) lookupRef(ref string) (*ir.Type, bool) {
	if t, ok := g.global.refs[ref]; ok {
		return t, true
	}
	if t, ok := g.local.refs[ref]; ok {
		return t, true
	}
	return nil, false
}

func (g *genctx) saveResponse(ref string, r *ir.StatusResponse) error {
	return g.local.saveResponse(ref, r)
}

func (g *genctx) saveWResponse(ref string, r *ir.StatusResponse) error {
	return g.local.saveWResponse(ref, r)
}

func (g *genctx) saveWType(ref string, t *ir.Type) error {
	return g.local.saveWType(ref, t)
}

func (g *genctx) lookupResponse(ref string) (*ir.StatusResponse, bool) {
	if r, ok := g.global.responses[ref]; ok {
		return r, true
	}
	if r, ok := g.local.responses[ref]; ok {
		return r, true
	}
	return nil, false
}

func (g *genctx) lookupWrappedResponse(ref string) (*ir.StatusResponse, bool) {
	if r, ok := g.global.wresponses[ref]; ok {
		return r, true
	}
	if r, ok := g.local.wresponses[ref]; ok {
		return r, true
	}
	return nil, false
}

func (g *genctx) lookupWType(ref string) (*ir.Type, bool) {
	if t, ok := g.global.wtypes[ref]; ok {
		return t, true
	}
	if t, ok := g.local.wtypes[ref]; ok {
		return t, true
	}
	return nil, false
}
