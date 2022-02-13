package gen

import (
	"github.com/go-faster/errors"
	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/jsonschema"
)

func (g *Generator) generateSchema(ctx *genctx, name string, schema *jsonschema.Schema) (*ir.Type, error) {
	gen := &schemaGen{
		localRefs: map[string]*ir.Type{},
		lookupRef: ctx.lookupRef,
	}

	t, err := gen.generate(name, schema)
	if err != nil {
		return nil, err
	}

	for _, t := range gen.side {
		if t.Is(ir.KindStruct) || (t.Is(ir.KindMap) && len(t.Fields) > 0) {
			if err := boxStructFields(ctx, t); err != nil {
				return nil, errors.Wrap(err, t.Name)
			}
		}

		if err := ctx.saveType(t); err != nil {
			return nil, errors.Wrap(err, "save inlined type")
		}
	}

	for ref, t := range gen.localRefs {
		if t.Is(ir.KindStruct) || (t.Is(ir.KindMap) && len(t.Fields) > 0) {
			if err := boxStructFields(ctx, t); err != nil {
				return nil, errors.Wrap(err, ref)
			}
		}
		if err := ctx.saveRef(ref, t); err != nil {
			return nil, errors.Wrap(err, "save referenced type")
		}
		// fmt.Println("saving ref", ref)
	}

	return t, nil
}
