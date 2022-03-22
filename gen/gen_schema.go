package gen

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"

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
	gen := &schemaGen{
		localRefs: map[string]*ir.Type{},
		lookupRef: ctx.lookupRef,
	}

	t, err := gen.generate(name, schema)
	if err != nil {
		return nil, err
	}

	if err := saveSchemaTypes(ctx, gen); err != nil {
		return nil, errors.Wrap(err, "save schema types")
	}

	return t, nil
}

type refResolver struct {
	root       []byte
	parsedRoot *jsonschema.RawSchema
}

func (r refResolver) parse(input []byte) (*jsonschema.RawSchema, error) {
	var rawSchema *jsonschema.RawSchema
	if err := json.Unmarshal(input, &rawSchema); err != nil {
		return nil, errors.Wrap(err, "unmarshal")
	}
	return rawSchema, nil
}

func (r refResolver) findPath(ref string, buf []byte) ([]byte, error) {
	for _, part := range strings.Split(ref, "/") {
		found := false
		if err := jx.DecodeBytes(buf).ObjBytes(func(d *jx.Decoder, key []byte) error {
			switch string(key) {
			case part:
				found = true
				raw, err := d.RawAppend(nil)
				if err != nil {
					return errors.Wrapf(err, "parse %q", key)
				}
				buf = raw
			default:
				return d.Skip()
			}
			return nil
		}); err != nil {
			return nil, err
		}

		if !found {
			return nil, errors.Errorf("find %q", part)
		}
	}
	return buf, nil
}

func (r refResolver) ResolveReference(ref string) (rawSchema *jsonschema.RawSchema, err error) {
	if !strings.HasPrefix(ref, "#") {
		return nil, errors.Errorf("unsupported ref %q", ref)
	}
	ref = strings.TrimPrefix(ref, "#")

	buf := r.root
	if !strings.ContainsRune(ref, '/') {
		if r.parsedRoot != nil {
			return r.parsedRoot, nil
		}
	} else {
		ref = strings.TrimPrefix(ref, "/")
		buf, err = r.findPath(ref, buf)
		if err != nil {
			return nil, errors.Wrapf(err, "find %q", ref)
		}
	}

	if err := json.Unmarshal(buf, &rawSchema); err != nil {
		return nil, errors.Wrap(err, "unmarshal")
	}
	return rawSchema, nil
}

// GenerateSchema generates type, validation and JSON encoding for given schema.
func GenerateSchema(input []byte, fs FileSystem, typeName, fileName, pkgName string) error {
	var rawSchema *jsonschema.RawSchema
	if err := json.Unmarshal(input, &rawSchema); err != nil {
		return errors.Wrap(err, "unmarshal")
	}

	p := jsonschema.NewParser(jsonschema.Settings{
		Resolver: refResolver{root: input},
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
	gen := &schemaGen{
		localRefs: map[string]*ir.Type{},
		lookupRef: func(ref string) (*ir.Type, bool) {
			return nil, false
		},
	}

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
