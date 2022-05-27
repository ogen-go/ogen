package gen

import (
	"bytes"
	"os"
	"strings"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/jsonschema"
)

func saveSchemaTypes(ctx *genctx, gen *schemaGen) error {
	for _, t := range gen.side {
		if t.IsStruct() {
			if err := boxStructFields(ctx, t); err != nil {
				return errors.Wrap(err, t.Name)
			}
		}

		if err := ctx.saveType(t); err != nil {
			return errors.Wrap(err, "save inlined type")
		}
	}

	for ref, t := range gen.localRefs {
		if t.IsStruct() {
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
	gen.log = g.log.Named("schemagen")

	t, err := gen.generate(name, schema)
	if err != nil {
		return nil, err
	}

	if err := saveSchemaTypes(ctx, gen); err != nil {
		return nil, errors.Wrap(err, "save schema types")
	}

	return t, nil
}

// GenerateSchemaOptions is options structure for GenerateSchema.
type GenerateSchemaOptions struct {
	// TypeName is root schema type name. Defaults to "Type".
	TypeName string
	// FileName is output filename. Defaults to "output.gen.go".
	FileName string
	// PkgName is the package name. Defaults to GOPACKAGE environment variable, if any. Otherwise, to "output".
	PkgName string
	// TrimPrefix is a ref name prefixes to trim. Defaults to []string{"#/definitions/", "#/$defs/"}.
	TrimPrefix []string
	// Logger to use.
	Logger *zap.Logger
}

func (o *GenerateSchemaOptions) setDefaults() {
	if o.TypeName == "" {
		o.TypeName = "Type"
	}
	if o.FileName == "" {
		o.FileName = "output.gen.go"
	}
	if o.PkgName == "" {
		o.PkgName = os.Getenv("GOPACKAGE")
		if o.PkgName == "" {
			o.PkgName = "output"
		}
	}
	if o.TrimPrefix == nil {
		o.TrimPrefix = []string{"#/definitions/", "#/$defs/"}
	}
	if o.Logger == nil {
		o.Logger = zap.NewNop()
	}
}

// GenerateSchema generates type, validation and JSON encoding for given schema.
func GenerateSchema(schema *jsonschema.Schema, fs FileSystem, opts GenerateSchemaOptions) error {
	opts.setDefaults()

	ctx := &genctx{
		jsonptr: newJSONPointer("#"),
		global:  newTStorage(),
		local:   newTStorage(),
	}

	gen := newSchemaGen(func(ref string) (*ir.Type, bool) {
		return nil, false
	})
	gen.log = opts.Logger.Named("schemagen")

	{
		prev := gen.nameRef
		gen.nameRef = func(ref string) (string, error) {
			for _, trim := range opts.TrimPrefix {
				ref = strings.TrimPrefix(ref, trim)
			}

			result, err := prev(ref)
			if err != nil {
				return "", err
			}
			return result, nil
		}
	}

	t, err := gen.generate(opts.TypeName, schema)
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
	if err := w.Generate("jsonschema", opts.FileName, TemplateConfig{
		Package: opts.PkgName,
		Types:   ctx.local.types,
	}); err != nil {
		return errors.Wrap(err, "write")
	}

	return nil
}
