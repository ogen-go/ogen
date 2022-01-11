package parser

import (
	"fmt"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/oas"
)

func (p *parser) parseSchema(schema *ogen.Schema) (*oas.Schema, error) {
	parser := &schemaParser{
		components: p.spec.Components.Schemas,
		globalRefs: p.refs.schemas,
		localRefs:  make(map[string]*oas.Schema),
	}

	s, err := parser.Parse(schema)
	if err != nil {
		return nil, errors.Wrap(err, "generate")
	}

	// Merge references.
	for ref, schema := range parser.localRefs {
		if _, found := p.refs.schemas[ref]; found {
			panic(fmt.Sprintf("schema reference conflict: %s", ref))
		}

		p.refs.schemas[ref] = schema
	}

	return s, nil
}
