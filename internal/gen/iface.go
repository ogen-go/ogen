package gen

func (g *Generator) implementRequest(s *Schema, m *Method) {
	var (
		ifaceName = m.Name + "Request"
		method    = "impl" + m.Name + "Request"
	)

	s.Implements[method] = struct{}{}
	iface, ok := g.interfaces[ifaceName]
	if !ok {
		iface = &Interface{
			Name:    ifaceName,
			Methods: map[string]struct{}{},
			Schemas: map[*Schema]struct{}{},
		}
	}

	iface.Methods[method] = struct{}{}
	iface.Schemas[s] = struct{}{}
	g.interfaces[ifaceName] = iface
}

func (g *Generator) unimplementRequest(s *Schema, m *Method) {
	var (
		ifaceName = m.Name + "Request"
		method    = "impl" + m.Name + "Request"
	)

	delete(s.Implements, method)
	iface, ok := g.interfaces[ifaceName]
	if !ok {
		panic("unreachable")
	}

	delete(iface.Schemas, s)
	if len(iface.Schemas) == 0 {
		delete(g.interfaces, ifaceName)
	}
}

func (g *Generator) implementResponse(r *Response, m *Method) {
	var (
		ifaceName = m.Name + "Response"
		method    = "impl" + m.Name + "Response"
	)

	iface, ok := g.interfaces[ifaceName]
	if !ok {
		iface = &Interface{
			Name:    ifaceName,
			Methods: map[string]struct{}{},
			Schemas: map[*Schema]struct{}{},
		}
	}

	iface.Methods[method] = struct{}{}
	g.interfaces[ifaceName] = iface
	for _, schema := range r.Contents {
		schema.Implements[method] = struct{}{}
		iface.Schemas[schema] = struct{}{}
	}
	if r.NoContent != nil {
		r.NoContent.Implements[method] = struct{}{}
	}
}

func (g *Generator) unimplementResponse(r *Response, m *Method) {
	var (
		ifaceName = m.Name + "Response"
		method    = "impl" + m.Name + "Response"
	)

	iface, ok := g.interfaces[ifaceName]
	if !ok {
		panic("unreachable")
	}

	for _, schema := range r.Contents {
		delete(schema.Implements, method)
		delete(iface.Schemas, schema)
	}
	if r.NoContent != nil {
		delete(r.NoContent.Implements, method)
	}

	if len(iface.Schemas) == 0 {
		delete(g.interfaces, ifaceName)
	}
}
