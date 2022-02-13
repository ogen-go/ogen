package gen

import "github.com/ogen-go/ogen/internal/ir"

// genctx is a generation context.
type genctx struct {
	path []string

	global *tstorage // readonly
	local  *tstorage
}

func newGenCtx(path ...string) *genctx {
	return &genctx{
		path:   path,
		global: newTStorage(),
		local:  newTStorage(),
	}
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

func (g *genctx) lookupWrappedType(ref string) (*ir.Type, bool) {
	if t, ok := g.global.wtypes[ref]; ok {
		return t, true
	}
	if t, ok := g.local.wtypes[ref]; ok {
		return t, true
	}
	return nil, false
}
