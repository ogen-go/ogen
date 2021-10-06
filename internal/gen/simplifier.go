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

	iface, ok := m.RequestType.(*ast.Interface)
	if !ok {
		return
	}

	for _, schema := range m.RequestBody.Contents {
		schema.Unimplement(iface)
		m.RequestType = schema

		if !m.RequestBody.Required {
			m.RequestType = &ast.Pointer{
				To: schema,
			}
		}
	}
}

func (g *Generator) devirtSingleResponse(m *ast.Method) {
	if len(m.Responses.StatusCode) != 1 || m.Responses.Default != nil {
		return
	}

	iface, ok := m.ResponseType.(*ast.Interface)
	if !ok {
		return
	}

	for _, resp := range m.Responses.StatusCode {
		if noc := resp.NoContent; noc != nil {
			resp.Unimplement(iface)
			m.ResponseType = noc
			continue
		}

		if len(resp.Contents) == 1 {
			resp.Unimplement(iface)
			for _, schema := range resp.Contents {
				m.ResponseType = schema
			}
		}
	}
}

func (g *Generator) devirtDefaultResponse(m *ast.Method) {
	if ok := (m.Responses.Default != nil && len(m.Responses.StatusCode) == 0); !ok {
		return
	}

	if len(m.Responses.Default.Contents) > 1 {
		return
	}

	iface, ok := m.ResponseType.(*ast.Interface)
	if !ok {
		return
	}

	m.Responses.Default.Unimplement(iface)
	m.ResponseType = m.Responses.Default.NoContent
}

func (g *Generator) removeUnusedIfaces() {
	for name, iface := range g.interfaces {
		if len(iface.Implementations) == 0 {
			delete(g.interfaces, name)
		}
	}
}
