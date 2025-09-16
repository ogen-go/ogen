package main

import (
	"encoding/json"
	"flag"
	"os"
	"slices"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen"
	"github.com/ogen-go/ogen/gen/ir"
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
		slices.SortStableFunc(r, func(a, b *ogen.Property) int {
			return strings.Compare(a.Name, b.Name)
		})
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
		// Any
		ogen.NewNamedSchema("Any", ogen.NewSchema()),
		// Empty struct
		ogen.NewNamedSchema("EmptyStruct", ogen.NewSchema().SetType("object")),
	}

	eachMapping(func(name, format, typ string) {
		do := func(s *ogen.Schema) {
			prefix := name
			if s.Nullable {
				prefix += "_nullable"
			}
			testSchemas = append(testSchemas,
				ogen.NewNamedSchema(prefix, s),
				ogen.NewNamedSchema(prefix+"_array", s.AsArray()),
				ogen.NewNamedSchema(prefix+"_array"+"_array", s.AsArray().AsArray()),
			)
		}
		do(ogen.NewSchema().
			SetType(typ).
			SetFormat(format))
		do(ogen.NewSchema().
			SetType(typ).
			SetFormat(format).SetNullable(true))
	})
	slices.SortStableFunc(testSchemas, func(a, b *ogen.NamedSchema) int {
		return strings.Compare(a.Name, b.Name)
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
						ir.EncodingJSON.String(): {Schema: &ogen.Schema{
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
						ir.EncodingJSON.String(): {
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
					ir.EncodingJSON.String(): {
						Schema: schema.Schema,
					},
				},
				Description: "Optional request body",
				Required:    false,
			})
		}},
		{"request_required", func(op *ogen.Operation, schema *ogen.NamedSchema) {
			op.SetRequestBody(&ogen.RequestBody{
				Content: map[string]ogen.Media{
					ir.EncodingJSON.String(): {
						Schema: schema.Schema,
					},
				},
				Description: "Required request body",
				Required:    true,
			})
		}},
		{"response", func(op *ogen.Operation, schema *ogen.NamedSchema) {
			op.SetResponses(ogen.Responses{
				"200": {
					Content: map[string]ogen.Media{
						ir.EncodingJSON.String(): {
							Schema: schema.Schema,
						},
					},
					Description: "Response",
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
						Ref: "#/components/responses/error",
					},
				}
			}

			spec.Paths["/"+name] = &ogen.PathItem{
				Post: op,
			}
		}
	}

	{
		op := ogen.NewOperation().SetOperationID("test_query_parameter").
			SetRequestBody(&ogen.RequestBody{
				Ref: "#/components/requestBodies/defaultBody",
			}).
			SetResponses(map[string]*ogen.Response{
				"200": {
					Ref: "#/components/responses/error",
				},
			})

		eachMapping(func(name, format, typ string) {
			if typ == "null" {
				return
			}

			add := func(name string, s *ogen.Schema) {
				op.Parameters = append(op.Parameters, &ogen.Parameter{
					Name:     name,
					In:       "query",
					Required: true,
					Schema:   s,
				})
			}

			s := &ogen.Schema{
				Type:   typ,
				Format: format,
			}
			add(name, s)
			add(name+"_array", s.AsArray())
		})
		p := op.Parameters
		slices.SortStableFunc(p, func(a, b *ogen.Parameter) int {
			return strings.Compare(a.Name, b.Name)
		})
		spec.Paths["/test_query_parameter"] = &ogen.PathItem{
			Post: op,
		}
	}

	return spec
}

func run() error {
	output := flag.String("output", "./_testdata/positive/test_format.json", "path to output file")
	flag.Parse()

	spec := generateSpec()

	data, err := json.MarshalIndent(spec, "", "\t")
	if err != nil {
		return errors.Wrap(err, "marshal spec")
	}

	return os.WriteFile(*output, data, 0o644)
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
