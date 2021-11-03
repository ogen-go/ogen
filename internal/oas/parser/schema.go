package parser

import (
	"fmt"

	"github.com/ogen-go/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/oas"
)

func (p *parser) parseSchema(schema ogen.Schema) (*oas.Schema, error) {
	gen := &schemaGen{
		spec:       p.spec,
		globalRefs: p.refs.schemas,
		localRefs:  make(map[string]*oas.Schema),
	}

	s, err := gen.Generate(schema)
	if err != nil {
		return nil, errors.Wrap(err, "generate")
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
