package parser

import (
	"fmt"

	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen"
	ast "github.com/ogen-go/ogen/internal/oas"
)

func (p *parser) parseSchema(schema ogen.Schema) (*ast.Schema, error) {
	gen := &schemaGen{
		spec:       p.spec,
		globalRefs: p.refs.schemas,
		localRefs:  make(map[string]*ast.Schema),
	}

	s, err := gen.Generate(schema)
	if err != nil {
		return nil, xerrors.Errorf("generate: %w", err)
	}

	// Merge references.
	for ref, schema := range gen.localRefs {
		if _, found := p.refs.schemas[ref]; found {
			panic(fmt.Sprintf("schema reference conflict: %s", ref))
		}

		p.refs.schemas[ref] = schema
	}

	return s, nil
}
