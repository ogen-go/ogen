package gen

func (g *Generator) simplify() {
	for _, method := range g.methods {
		if method.RequestBody != nil {
			switch len(method.RequestBody.Contents) {
			case 0:
			case 1:
				g.devirtSingleRequest(method)
			default:
				g.devirtManyEqualRequests(method)
			}
		}

		switch len(method.Responses) {
		case 0:
		case 1:
			g.devirtSingleResponse(method)
		}
	}

	g.removeUnusedIfaces()
}

// devirtSingleRequest removes interface in case
// where method have only one content in requestBody.
func (g *Generator) devirtSingleRequest(m *Method) {
	if len(m.RequestBody.Contents) != 1 {
		return
	}

	if iface, ok := g.interfaces[m.RequestType]; ok {
		for _, schema := range m.RequestBody.Contents {
			schema.unimplement(iface)
			m.RequestType = "*" + schema.Name
		}
	}
}

// devirtManyEqualRequests removes interface
// and squashes all request types into a single struct
// if all schemas in different content-types have the same fields.
func (g *Generator) devirtManyEqualRequests(m *Method) {
	if len(m.RequestBody.Contents) < 2 {
		return
	}

	iface, ok := g.interfaces[m.RequestType]
	if !ok {
		return
	}

	var schemas []*Schema
	for _, schema := range m.RequestBody.Contents {
		schemas = append(schemas, schema)
	}

	root := schemas[0]
	for _, s := range schemas[1:] {
		if !root.EqualFields(*s) {
			return
		}
	}

	for _, s := range schemas {
		s.unimplement(iface)
	}

	m.RequestType = "*" + root.Name
	for contentType := range m.RequestBody.Contents {
		m.RequestBody.Contents[contentType] = root
	}
}

func (g *Generator) devirtSingleResponse(m *Method) {
	if len(m.Responses) != 1 || m.ResponseDefault != nil {
		return
	}

	if iface, ok := g.interfaces[m.ResponseType]; ok {
		for _, resp := range m.Responses {
			if len(resp.Contents) == 1 {
				resp.unimplement(iface)
				for _, schema := range resp.Contents {
					m.ResponseType = "*" + schema.Name
				}
			}
		}
	}
}

func (g *Generator) removeUnusedIfaces() {
	for name, iface := range g.interfaces {
		if len(iface.Implementations) == 0 {
			delete(g.interfaces, name)
		}
	}
}
