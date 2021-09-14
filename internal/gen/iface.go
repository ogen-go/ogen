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
