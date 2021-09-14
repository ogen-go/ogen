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
	}
}

// devirtSingleRequest removes interface in case
// where method have only one content in requestBody.
func (g *Generator) devirtSingleRequest(m *Method) {
	if len(m.RequestBody.Contents) != 1 {
		return
	}

	for _, schema := range m.RequestBody.Contents {
		g.unimplementRequest(schema, m)
		m.RequestType = "*" + schema.Name
	}
}

// devirtManyEqualRequests removes interface
// and squashes all request types into a single struct
// if all schemas in different content-types have the same fields.
func (g *Generator) devirtManyEqualRequests(m *Method) {
	if len(m.RequestBody.Contents) < 2 {
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
		g.unimplementRequest(s, m)
		delete(g.schemas, s.Name)
	}

	root.Name = m.Name + "Request"
	g.schemas[root.Name] = root

	m.RequestType = "*" + root.Name
	for contentType := range m.RequestBody.Contents {
		m.RequestBody.Contents[contentType] = root
	}
}
