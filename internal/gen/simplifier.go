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
	// TODO(ernado): Remove unused generics
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

	for contentType, schema := range m.RequestBody.Contents {
		schema.Unimplement(iface)
		schema, unwrapped := g.unwrapAlias(schema)
		if unwrapped {
			m.RequestBody.Contents[contentType] = schema
		}

		m.RequestType = schema
		if !m.RequestBody.Required {
			m.RequestType = ast.Pointer(schema, ast.NilInvalid)
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
			noc, unwrapped := g.unwrapAlias(noc)
			if unwrapped {
				resp.NoContent = noc
			}

			m.ResponseType = noc
			continue
		}

		if len(resp.Contents) == 1 {
			resp.Unimplement(iface)
			for ctype, schema := range resp.Contents {
				schema, unwrapped := g.unwrapAlias(schema)
				if unwrapped {
					resp.Contents[ctype] = schema
				}

				m.ResponseType = schema
			}
		}
	}
}

func (g *Generator) devirtDefaultResponse(m *ast.Method) {
	if !(m.Responses.Default != nil && len(m.Responses.StatusCode) == 0) {
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
	if noc := m.Responses.Default.NoContent; noc != nil {
		m.ResponseType = noc
		return
	}

	for _, schema := range m.Responses.Default.Contents {
		m.ResponseType = schema
		return
	}
}

func (g *Generator) removeUnusedIfaces() {
	for name, iface := range g.interfaces {
		if len(iface.Implementations) == 0 {
			delete(g.interfaces, name)
		}
	}
}

func (g *Generator) unwrapAlias(schema *ast.Schema) (*ast.Schema, bool) {
	if !schema.Is(ast.KindAlias) {
		return schema, false
	}

	to := schema.AliasTo
	if to.Is(ast.KindPrimitive) {
		if to.Primitive == ast.EmptyStruct {
			return schema, false
		}
	}

	delete(g.schemas, schema.Name)
	return schema.AliasTo, true
}
