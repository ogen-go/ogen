package parser

import (
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"testing"

	"github.com/go-faster/yaml"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
)

var testRootURL = &url.URL{
	Path: "/root.json",
}

func compareJSON(t require.TestingT, expected, got any) {
	// Compare as JSON because we can't skip locators.
	encode := func(s any) string {
		b, err := json.MarshalIndent(s, "", "  ")
		require.NoError(t, err)
		return string(b)
	}
	e, g := encode(expected), encode(got)
	require.JSONEq(t, e, g)
}

func TestComplicatedReference(t *testing.T) {
	root := &ogen.Spec{
		OpenAPI: "3.0.3",
		Paths: map[string]*ogen.PathItem{
			"/get": {
				Get: &ogen.Operation{
					OperationID: "testGet",
					Description: "operation description",
					Parameters: []*ogen.Parameter{
						{Ref: "#/paths/~1post/post/parameters/0"},
					},
					RequestBody: &ogen.RequestBody{
						Ref: "#/paths/~1post/post/requestBody",
					},
					Responses: map[string]*ogen.Response{
						"200": {
							Ref: "#/paths/~1post/post/responses/200",
						},
						"201": {
							Headers: map[string]*ogen.Header{
								"ResponseHeader": {
									Schema: &ogen.Schema{Type: "string"},
									Style:  "simple",
								},
							},
							Content: map[string]ogen.Media{
								"application/json": {
									Schema: &ogen.Schema{Type: "string"},
								},
							},
						},
					},
				},
			},
			"/post": {
				Post: &ogen.Operation{
					OperationID: "testPost",
					Description: "operation description",
					Parameters: []*ogen.Parameter{
						{
							Name:  "param",
							In:    "query",
							Style: "form",
							Schema: &ogen.Schema{
								Type: "string",
							},
						},
					},
					RequestBody: &ogen.RequestBody{
						Content: map[string]ogen.Media{
							"application/json": {
								Schema: &ogen.Schema{
									Type: "string",
								},
							},
						},
					},
					Responses: map[string]*ogen.Response{
						"200": {
							Headers: map[string]*ogen.Header{
								"ResponseHeader": {
									Schema: &ogen.Schema{Type: "string"},
									Style:  "simple",
								},
							},
							Content: map[string]ogen.Media{
								"application/json": {
									Schema: &ogen.Schema{
										Ref: "#/paths/~1post/post/requestBody/content/application~1json/schema",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	a := require.New(t)

	var raw yaml.Node
	a.NoError(raw.Encode(root))
	root.Raw = &raw

	spec, err := Parse(root, Settings{
		RootURL: testRootURL,
	})
	a.NoError(err)

	var (
		responseHeader = &openapi.Header{
			Name:   "ResponseHeader",
			Schema: &jsonschema.Schema{Type: "string"},
			In:     openapi.LocationHeader,
			Style:  openapi.HeaderStyleSimple,
		}
		testGet = &openapi.Operation{
			OperationID: "testGet",
			Description: "operation description",
			HTTPMethod:  "get",
			Path: openapi.Path{
				{Raw: "/get"},
			},
			Parameters: []*openapi.Parameter{
				{
					Ref:  refKey{Loc: "/root.json", Ptr: "#/paths/~1post/post/parameters/0"},
					Name: "param",
					Schema: &jsonschema.Schema{
						Type: "string",
					},
					In:      openapi.LocationQuery,
					Style:   openapi.QueryStyleForm,
					Explode: true,
				},
			},
			RequestBody: &openapi.RequestBody{
				Ref: refKey{Loc: "/root.json", Ptr: "#/paths/~1post/post/requestBody"},
				Content: map[string]*openapi.MediaType{
					"application/json": {
						Schema: &jsonschema.Schema{
							Type: "string",
						},
						Examples: map[string]*openapi.Example{},
						Encoding: map[string]*openapi.Encoding{},
					},
				},
			},
			Security: []openapi.SecurityRequirement{},
			Responses: openapi.Responses{
				StatusCode: map[int]*openapi.Response{
					200: {
						Ref: refKey{Loc: "/root.json", Ptr: "#/paths/~1post/post/responses/200"},
						Headers: map[string]*openapi.Header{
							"ResponseHeader": responseHeader,
						},
						Content: map[string]*openapi.MediaType{
							"application/json": {
								Schema: &jsonschema.Schema{
									Ref:  refKey{Loc: "/root.json", Ptr: "#/paths/~1post/post/requestBody/content/application~1json/schema"},
									Type: "string",
								},
								Examples: map[string]*openapi.Example{},
								Encoding: map[string]*openapi.Encoding{},
							},
						},
					},
					201: {
						Headers: map[string]*openapi.Header{
							"ResponseHeader": responseHeader,
						},
						Content: map[string]*openapi.MediaType{
							"application/json": {
								Schema:   &jsonschema.Schema{Type: "string"},
								Examples: map[string]*openapi.Example{},
								Encoding: map[string]*openapi.Encoding{},
							},
						},
					},
				},
			},
		}
		testPost = &openapi.Operation{
			OperationID: "testPost",
			Description: "operation description",
			HTTPMethod:  "post",
			Path: openapi.Path{
				{Raw: "/post"},
			},
			Parameters: []*openapi.Parameter{
				{
					Name: "param",
					Schema: &jsonschema.Schema{
						Type: "string",
					},
					In:      openapi.LocationQuery,
					Style:   openapi.QueryStyleForm,
					Explode: true,
				},
			},
			RequestBody: &openapi.RequestBody{
				Content: map[string]*openapi.MediaType{
					"application/json": {
						Schema: &jsonschema.Schema{
							Type: "string",
						},
						Examples: map[string]*openapi.Example{},
						Encoding: map[string]*openapi.Encoding{},
					},
				},
			},
			Security: []openapi.SecurityRequirement{},
			Responses: openapi.Responses{
				StatusCode: map[int]*openapi.Response{
					200: {
						Headers: map[string]*openapi.Header{
							"ResponseHeader": responseHeader,
						},
						Content: map[string]*openapi.MediaType{
							"application/json": {
								Schema: &jsonschema.Schema{
									Ref:  refKey{Loc: "/root.json", Ptr: "#/paths/~1post/post/requestBody/content/application~1json/schema"},
									Type: "string",
								},
								Examples: map[string]*openapi.Example{},
								Encoding: map[string]*openapi.Encoding{},
							},
						},
					},
				},
			},
		}
	)
	slices.SortFunc(spec.Operations, func(a, b *openapi.Operation) int {
		return strings.Compare(a.OperationID, b.OperationID)
	})
	compareJSON(t, &openapi.API{
		Version: openapi.Version{
			Major: 3,
			Minor: 0,
			Patch: 3,
		},
		Operations: []*openapi.Operation{
			testGet,
			testPost,
		},
		Components: &openapi.Components{
			Schemas:       map[string]*jsonschema.Schema{},
			Responses:     map[string]*openapi.Response{},
			Parameters:    map[string]*openapi.Parameter{},
			Examples:      map[string]*openapi.Example{},
			RequestBodies: map[string]*openapi.RequestBody{},
		},
	}, spec)
}

func TestParserNoPanic(t *testing.T) {
	schema := func(s *ogen.Schema) *ogen.Spec {
		return &ogen.Spec{
			Components: &ogen.Components{
				Schemas: map[string]*ogen.Schema{
					"schema": s,
				},
			},
		}
	}

	inputs := []*ogen.Spec{
		nil,
		{},
		{
			Paths: ogen.Paths{},
		},
		{
			Components: &ogen.Components{},
		},
		{
			Components: &ogen.Components{
				Examples: map[string]*ogen.Example{
					"example": nil,
				},
			},
		},
		schema(nil),
		schema(&ogen.Schema{}),
		schema(&ogen.Schema{
			AllOf: []*ogen.Schema{nil},
		}),
		schema(&ogen.Schema{
			OneOf: []*ogen.Schema{nil},
		}),
		schema(&ogen.Schema{
			AnyOf: []*ogen.Schema{nil},
		}),
		schema(&ogen.Schema{
			Type: "array",
		}),
		schema(&ogen.Schema{
			Type: "object",
			Properties: ogen.Properties{
				{Name: "foo", Schema: nil},
			},
		}),
		schema(&ogen.Schema{
			Type: "object",
			PatternProperties: ogen.PatternProperties{
				{Pattern: "foo", Schema: nil},
			},
		}),
	}
	for i, tt := range inputs {
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			require.NotPanics(t, func() {
				_, _ = Parse(tt, Settings{})
			})
			require.NotPanics(t, func() {
				_, _ = Parse(tt, Settings{
					InferTypes: true,
				})
			})
		})
	}
}
