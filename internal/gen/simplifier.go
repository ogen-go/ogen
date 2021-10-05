package gen

import "github.com/ogen-go/ogen/internal/ast"

func (g *Generator) simplify() {
	for _, method := range g.methods {
		if method.RequestBody != nil {
			if len(method.RequestBody.Contents) == 1 {
				g.devirtSingleRequest(method)
			}
		}

		g.devirtDefaultResponse(method)
		g.devirtSingleResponse(method)
	}

	g.removeUnusedIfaces()
}

// devirtSingleRequest removes interface in case
// where method have only one content in requestBody.
func (g *Generator) devirtSingleRequest(m *ast.Method) {
	if len(m.RequestBody.Contents) != 1 {
		return
	}

	if iface, ok := g.interfaces[m.RequestType]; ok {
		for _, schema := range m.RequestBody.Contents {
			schema.Unimplement(iface)
			m.RequestType = "*" + schema.Type()
		}
	}
}

func (g *Generator) devirtSingleResponse(m *ast.Method) {
	if len(m.Responses) != 1 || m.ResponseDefault != nil {
		return
	}

	if iface, ok := g.interfaces[m.ResponseType]; ok {
		for _, resp := range m.Responses {
			if noc := resp.NoContent; noc != nil {
				resp.Unimplement(iface)
				m.ResponseType = "*" + noc.Type()
				continue
			}

			if len(resp.Contents) == 1 {
				resp.Unimplement(iface)
				for _, schema := range resp.Contents {
					m.ResponseType = "*" + schema.Type()
				}
			}
		}
	}
}

func (g *Generator) devirtDefaultResponse(m *ast.Method) {
	if ok := (m.ResponseDefault != nil && len(m.Responses) == 0); !ok {
		return
	}

	if len(m.ResponseDefault.Contents) > 1 {
		return
	}

	if iface, ok := g.interfaces[m.ResponseType]; ok {
		m.ResponseDefault.Unimplement(iface)
		m.ResponseType = "*" + m.ResponseDefault.NoContent.Type()
	}
}

func (g *Generator) removeUnusedIfaces() {
	for name, iface := range g.interfaces {
		if len(iface.Implementations) == 0 {
			delete(g.interfaces, name)
		}
	}
}
