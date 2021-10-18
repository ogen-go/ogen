package gen

import (
	"fmt"

	"golang.org/x/xerrors"

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
		return nil, xerrors.Errorf("generate: %w", err)
	}

	// Merge nested objects.
	for _, side := range gen.side {
		if side.Is(ast.KindPrimitive, ast.KindArray, ast.KindPointer) {
			panic("unreachable")
		}

		if _, found := g.schemas[side.Name]; found && !side.IsGeneric() {
			panic(fmt.Sprintf("side schema name conflict: %s", side.Name))
		}

		g.schemas[side.Name] = side
	}

	// Merge references.
	for ref, schema := range gen.localRefs {
		if schema.Is(ast.KindPrimitive, ast.KindArray, ast.KindPointer) {
			panic("unreachable")
		}
		if _, found := g.schemaRefs[ref]; found {
			panic(fmt.Sprintf("schema reference conflict: %s", ref))
		}
		if _, found := g.schemas[schema.Name]; found {
			panic(fmt.Sprintf("schema reference name conflict: %s", ref))
		}

		g.schemaRefs[ref] = schema
		g.schemas[schema.Name] = schema
	}

	return s, nil
}
