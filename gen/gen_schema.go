package gen

import (
	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/internal/jsonschema"
)

func (g *Generator) generateSchema(name string, schema *jsonschema.Schema) (*ir.Type, error) {
	gen := &schemaGen{
		localRefs:  map[string]*ir.Type{},
		globalRefs: g.refs.types,
	}

	t, err := gen.generate(name, schema)
	if err != nil {
		return nil, err
	}

	for _, side := range gen.side {
		g.saveType(side)
	}

	for ref, t := range gen.localRefs {
		g.saveRef(ref, t)
	}

	return t, nil
}
