package parser

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/jsonpointer"
	"github.com/ogen-go/ogen/jsonschema"
)

func (p *parser) parseSchema(schema *ogen.Schema, ctx *jsonpointer.ResolveCtx) (*jsonschema.Schema, error) {
	s, err := p.schemaParser.ParseWithContext(schema.ToJSONSchema(), ctx)
	if err != nil {
		return nil, errors.Wrap(err, "parse schema")
	}
	return s, nil
}
