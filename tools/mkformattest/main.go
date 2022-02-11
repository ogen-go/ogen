package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen"
	"github.com/ogen-go/ogen/internal/ir"
)

func generateSpec() *ogen.Spec {
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

	return spec
}

func run() error {
	output := flag.String("output", "./_testdata/test_format.json", "path to output file")
	flag.Parse()

	spec := generateSpec()

	f, err := os.Create(*output)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	return json.NewEncoder(f).Encode(spec)
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
