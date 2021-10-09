package gen

import (
	"fmt"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/ast"
)

func (g *Generator) generateSchema(name string, schema ogen.Schema) (*ast.Schema, error) {
	gen := &schemaGen{
		spec:       g.spec,
		globalRefs: g.schemaRefs,
		localRefs:  make(map[string]*ast.Schema),
	}

	s, err := gen.Generate(name, schema)
	if err != nil {
		return nil, err
	}

	for _, side := range gen.side {
		if side.Is(ast.KindPrimitive, ast.KindArray) {
			panic("unreachable")
		}

		if _, found := g.schemas[side.Name]; found {
			panic(fmt.Sprintf("side schema name conflict: %s", side.Name))
		}

		g.schemas[side.Name] = side
	}

	for ref, schema := range gen.localRefs {
		g.schemaRefs[ref] = schema
	}

	return s, nil
}
