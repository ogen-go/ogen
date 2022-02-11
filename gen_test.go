package ogen_test

import (
	"embed"
	"go/format"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen"
	"github.com/ogen-go/ogen/internal/ir"
)

//go:embed _testdata
var testdata embed.FS

// TODO: Create validationFs.
type fmtFs struct{}

func (n fmtFs) WriteFile(baseName string, source []byte) error {
	_, err := format.Source(source)
	return err
}

func testGenerate(t *testing.T, name string, ignore ...string) {
	t.Helper()

	data, err := testdata.ReadFile(name)
	require.NoError(t, err)
	spec, err := ogen.Parse(data)
	require.NoError(t, err)
	opt := gen.Options{
		IgnoreNotImplemented: ignore,
		InferSchemaType:      true,
	}
	t.Run("Gen", func(t *testing.T) {
		defer func() {
			if rr := recover(); rr != nil {
				t.Fatalf("panic: %+v", rr)
			}
		}()

		g, err := gen.NewGenerator(spec, opt)
		require.NoError(t, err)

		require.NoError(t, g.WriteSource(fmtFs{}, "api"))
	})
	if len(opt.IgnoreNotImplemented) > 0 {
		t.Run("Full", func(t *testing.T) {
			t.Skipf("Ignoring: %s", opt.IgnoreNotImplemented)
		})
	}
}

func TestGenerate(t *testing.T) {
	t.Parallel()
	g := func(name string, ignore ...string) func(t *testing.T) {
		return func(t *testing.T) {
			t.Helper()
			t.Parallel()
			testGenerate(t, name, ignore...)
		}
	}

	skipSets := map[string][]string{
		"petstore.yaml": {},
		"petstore-expanded.yaml": {
			"allOf",
		},
		"firecracker.json": {},
		"sample.json":      {},
		"manga.json": {
			"unsupported content types",
		},
		"techempower.json": {},
		"telegram_bot_api.json": {
			"anyOf",
			"unsupported content types",
		},
		"gotd_bot_api.json": {
			"unsupported content types",
		},
		"k8s.json": {
			"unsupported content types",
		},
		"api.github.com.json": {
			"complex parameter types",
			"complex anyOf",
			"allOf",
			"discriminator inference",
			"sum types with same names",
			"sum type parameter",
			"unsupported content types",
			"empty schema",
		},
		"tinkoff.json": {},
	}

	if err := fs.WalkDir(testdata, "_testdata", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return err
		}
		_, file := filepath.Split(path)

		skip, ok := skipSets[file]
		if !ok {
			skip = []string{"all"}
		}
		t.Run(strings.TrimSuffix(file, ".json"), g(path, skip...))
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestFormats(t *testing.T) {
	// expected result
	mapping := gen.TypeFormatMapping()
	eachMapping := func(cb func(name, format, typ string)) {
		for typ, m := range mapping {
			for typeFormat := range m {
				name := string(typ)
				if typeFormat != "" {
					name += "_" + typeFormat
				}
				cb(name, typeFormat, string(typ))
			}
		}
	}

	properties := func(prefix string) (r []*ogen.Property) {
		add := func(name string, s *ogen.Schema) {
			for _, subs := range []struct {
				name   string
				schema *ogen.Schema
			}{
				{"_", s},
				{"_array_", s.AsArray()},
				{"_double_array_", s.AsArray().AsArray()},
			} {
				r = append(r,
					subs.schema.ToProperty(prefix+subs.name+name),
				)
			}
		}

		eachMapping(func(name, format, typ string) {
			add(name, ogen.NewSchema().
				SetType(typ).
				SetFormat(format))
		})

		add("any", ogen.NewSchema())
		return r
	}

	testSchemas := []*ogen.NamedSchema{
		ogen.NewNamedSchema("FormatTest", ogen.NewSchema().
			AddRequiredProperties(
				properties("required")...,
			).
			AddOptionalProperties(
				properties("optional")...,
			)),
		// Empty
		ogen.NewNamedSchema("Any", ogen.NewSchema()),
	}

	eachMapping(func(name, format, typ string) {
		do := func(s *ogen.Schema) {
			prefix := name
			if s.Nullable {
				prefix += "_nullable"
			}
			testSchemas = append(testSchemas, ogen.NewNamedSchema(prefix, s))
			testSchemas = append(testSchemas, ogen.NewNamedSchema(prefix+"_array", s.AsArray()))
			testSchemas = append(testSchemas, ogen.NewNamedSchema(prefix+"_array"+"_array", s.AsArray().AsArray()))
		}
		do(ogen.NewSchema().
			SetType(typ).
			SetFormat(format))
		do(ogen.NewSchema().
			SetType(typ).
			SetFormat(format).SetNullable(true))
	})

	spec := &ogen.Spec{
		OpenAPI: "3.1.0",
		Info: ogen.Info{
			Title:       "Format test",
			Description: "Auto-generated testing schema for checking format and primitive encoding and decoding",
			Version:     "0.1.0",
		},
		Paths: map[string]*ogen.PathItem{},
		Components: &ogen.Components{
			Responses: ogen.Responses{
				"error": {
					Description: "An Error Response",
					Content: map[string]ogen.Media{
						string(ir.ContentTypeJSON): {Schema: &ogen.Schema{
							Type:        "object",
							Description: "Error Response Schema",
							Properties: []ogen.Property{
								{Name: "code", Schema: &ogen.Schema{Type: "integer", Format: "int32"}},
								{Name: "status", Schema: &ogen.Schema{Type: "string"}},
							},
						}},
					},
				},
			},
			RequestBodies: map[string]*ogen.RequestBody{
				"defaultBody": {
					Description: "Referenced RequestBody",
					Content: map[string]ogen.Media{
						string(ir.ContentTypeJSON): {
							Schema: &ogen.Schema{Type: "string"},
						},
					},
					Required: true,
				},
			},
		},
	}

	type placement struct {
		name  string
		apply func(op *ogen.Operation, schema *ogen.NamedSchema)
	}
	for _, p := range []placement{
		{"request", func(op *ogen.Operation, schema *ogen.NamedSchema) {
			op.SetRequestBody(&ogen.RequestBody{
				Content: map[string]ogen.Media{
					string(ir.ContentTypeJSON): {
						Schema: schema.Schema,
					},
				},
				Required: false,
			})
		}},
		{"response", func(op *ogen.Operation, schema *ogen.NamedSchema) {
			op.SetResponses(ogen.Responses{
				"200": {
					Content: map[string]ogen.Media{
						string(ir.ContentTypeJSON): {
							Schema: schema.Schema,
						},
					},
				},
			})
		}},
	} {
		for _, s := range testSchemas {
			name := "test_" + p.name + "_" + s.Name

			op := ogen.NewOperation().SetOperationID(name)
			p.apply(op, s)
			if op.RequestBody == nil {
				op.RequestBody = &ogen.RequestBody{
					Ref: "#/components/requestBodies/defaultBody",
				}
			}
			if len(op.Responses) == 0 {
				op.Responses = map[string]*ogen.Response{
					"200": {
						Ref:         "#/components/responses/error",
						Description: "description",
					},
				}
			}

			spec.Paths["/"+name] = &ogen.PathItem{
				Post: op,
			}
		}
	}

	g, err := gen.NewGenerator(spec, gen.Options{
		InferSchemaType: true,
	})
	require.NoError(t, err)
	require.NoError(t, g.WriteSource(fmtFs{}, "api"))
}
