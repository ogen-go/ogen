package gen

import (
	"os"
	"strings"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/internal/xmaps"
	"github.com/ogen-go/ogen/jsonschema"
)

func saveSchemaTypes(ctx *genctx, gen *schemaGen, refEncoding map[jsonschema.Ref]ir.Encoding) error {
	for _, t := range gen.side {
		if err := ctx.saveType(t); err != nil {
			return errors.Wrap(err, "save inlined type")
		}
	}

	for ref, t := range gen.localRefs {
		encoding := ir.EncodingJSON
		if e, ok := refEncoding[ref]; ok {
			encoding = e
		}
		if err := ctx.saveRef(ref, encoding, t); err != nil {
			return errors.Wrap(err, "save referenced type")
		}
	}
	return nil
}

type generateSchemaOverride struct {
	refEncoding map[jsonschema.Ref]ir.Encoding
	nameRef     func(ref jsonschema.Ref, def refNamer) (string, error)
	fieldMut    func(*ir.Field) error
}

func (g *Generator) generateSchema(
	ctx *genctx,
	name string,
	schema *jsonschema.Schema,
	optional bool,
	override *generateSchemaOverride,
) (_ *ir.Type, rerr error) {
	defer handleSchemaDepth(schema, &rerr)

	lookup := func(ref jsonschema.Ref) (*ir.Type, bool) {
		encoding := ir.EncodingJSON
		if o := override; o != nil {
			if e, ok := o.refEncoding[ref]; ok {
				encoding = e
			}
		}
		return ctx.lookupRef(ref, encoding)
	}

	gen := newSchemaGen(lookup)
	if o := override; o != nil {
		if n := o.nameRef; n != nil {
			prev := gen.nameRef
			gen.nameRef = func(ref jsonschema.Ref) (string, error) {
				return n(ref, prev)
			}
		}
		if m := o.fieldMut; m != nil {
			gen.fieldMut = m
		}
	}
	gen.log = g.log.Named("schemagen")
	gen.fail = g.fail
	gen.depthLimit = g.parseOpts.SchemaDepthLimit
	gen.imports = g.imports

	t, err := gen.generate(name, schema, optional)
	if err != nil {
		return nil, err
	}

	var refEncoding map[jsonschema.Ref]ir.Encoding
	if o := override; o != nil {
		refEncoding = o.refEncoding
	}
	if err := saveSchemaTypes(ctx, gen, refEncoding); err != nil {
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
func GenerateSchema(schema *jsonschema.Schema, fs FileSystem, opts GenerateSchemaOptions) (rerr error) {
	defer handleSchemaDepth(schema, &rerr)

	opts.setDefaults()

	ctx := &genctx{
		global: newTStorage(),
		local:  newTStorage(),
	}

	// TODO(tdakkota): pass input filename
	gen := newSchemaGen(func(ref jsonschema.Ref) (*ir.Type, bool) {
		return nil, false
	})
	gen.log = opts.Logger.Named("schemagen")

	{
		prev := gen.nameRef
		gen.nameRef = func(ref jsonschema.Ref) (string, error) {
			nref := ref
			for _, trim := range opts.TrimPrefix {
				nref.Ptr = strings.TrimPrefix(nref.Ptr, trim)
			}
			if !strings.HasPrefix(nref.Ptr, "#") {
				nref.Ptr = "#" + nref.Ptr
			}

			result, err := prev(nref)
			if err != nil {
				return "", err
			}
			return result, nil
		}
	}

	t, err := gen.generate(opts.TypeName, schema, false)
	if err != nil {
		return errors.Wrap(err, "generate type")
	}
	t.AddFeature("json")

	if err := saveSchemaTypes(ctx, gen, nil); err != nil {
		return errors.Wrap(err, "save schema types")
	}

	types := ctx.local.types
	for _, key := range xmaps.SortedKeys(types) {
		if t := types[key]; t.IsStruct() {
			if err := checkStructRecursions(t); err != nil {
				return errors.Wrap(err, t.Name)
			}
		}
	}

	w := &writer{
		fs: fs,
		t:  vendoredTemplates(),
	}
	if err := w.Generate("jsonschema", opts.FileName, TemplateConfig{
		Package: opts.PkgName,
		Types:   ctx.local.types,
		Imports: gen.imports,
	}); err != nil {
		return errors.Wrap(err, "write")
	}

	return nil
}
