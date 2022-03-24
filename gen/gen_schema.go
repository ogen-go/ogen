package gen

import (
	"bytes"
	"encoding/json"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/jsonschema"
)

func saveSchemaTypes(ctx *genctx, gen *schemaGen) error {
	for _, t := range gen.side {
		if t.Is(ir.KindStruct) || (t.Is(ir.KindMap) && len(t.Fields) > 0) {
			if err := boxStructFields(ctx, t); err != nil {
				return errors.Wrap(err, t.Name)
			}
		}

		if err := ctx.saveType(t); err != nil {
			return errors.Wrap(err, "save inlined type")
		}
	}

	for ref, t := range gen.localRefs {
		if t.Is(ir.KindStruct) || (t.Is(ir.KindMap) && len(t.Fields) > 0) {
			if err := boxStructFields(ctx, t); err != nil {
				return errors.Wrap(err, ref)
			}
		}
		if err := ctx.saveRef(ref, t); err != nil {
			return errors.Wrap(err, "save referenced type")
		}
	}
	return nil
}

func (g *Generator) generateSchema(ctx *genctx, name string, schema *jsonschema.Schema) (*ir.Type, error) {
	gen := newSchemaGen(ctx.lookupRef)

	t, err := gen.generate(name, schema)
	if err != nil {
		return nil, err
	}

	if err := saveSchemaTypes(ctx, gen); err != nil {
		return nil, errors.Wrap(err, "save schema types")
	}

	return t, nil
}

// GenerateSchema generates type, validation and JSON encoding for given schema.
func GenerateSchema(input []byte, fs FileSystem, typeName, fileName, pkgName string) error {
	var rawSchema *jsonschema.RawSchema
	if err := json.Unmarshal(input, &rawSchema); err != nil {
		return errors.Wrap(err, "unmarshal")
	}

	p := jsonschema.NewParser(jsonschema.Settings{
		Resolver: jsonschema.NewRootResolver(input),
	})
	schema, err := p.Parse(rawSchema)
	if err != nil {
		return errors.Wrap(err, "parse")
	}

	ctx := &genctx{
		path:   []string{"#"},
		global: newTStorage(),
		local:  newTStorage(),
	}
	gen := newSchemaGen(func(ref string) (*ir.Type, bool) {
		return nil, false
	})

	t, err := gen.generate(typeName, schema)
	if err != nil {
		return errors.Wrap(err, "generate type")
	}
	t.AddFeature("json")

	if err := saveSchemaTypes(ctx, gen); err != nil {
		return errors.Wrap(err, "save schema types")
	}

	w := &writer{
		fs:    fs,
		t:     vendoredTemplates(),
		buf:   new(bytes.Buffer),
		wrote: map[string]bool{},
	}
	if err := w.Generate("jsonschema", fileName, TemplateConfig{
		Package: pkgName,
		Types:   ctx.local.types,
	}); err != nil {
		return errors.Wrap(err, "write")
	}

	return nil
}
