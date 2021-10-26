package gen

import (
	"github.com/ogen-go/ogen/internal/ast"
	"github.com/ogen-go/ogen/internal/ir"
)

func (g *Generator) generateSchema(name string, schema *ast.Schema) (*ir.Type, error) {
	gen := &schemaGen{
		localRefs:  map[string]*ir.Type{},
		globalRefs: g.refs,
	}

	typ, err := gen.generate(name, schema)
	if err != nil {
		return nil, err
	}

	for _, side := range gen.side {
		g.saveType(side)
	}

	for ref, typ := range gen.localRefs {
		g.saveRef(ref, typ)
	}

	return typ, nil
}
