package parser

import (
	"context"
	"encoding/json"
	"sort"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
)

type external map[string]interface{}

func (e external) Get(_ context.Context, loc string) ([]byte, error) {
	r, ok := e[loc]
	if !ok {
		return nil, errors.Errorf("unexpected location %q", loc)
	}
	return json.Marshal(r)
}

func TestExternalReference(t *testing.T) {
	exampleValue := json.RawMessage(`"value"`)

	root := &ogen.Spec{
		Paths: map[string]*ogen.PathItem{
			"/get": {
				Get: &ogen.Operation{
					OperationID: "testGet",
					Description: "operation description",
					Parameters: []*ogen.Parameter{
						{Ref: "#/components/parameters/LocalParameter"},
					},
					RequestBody: &ogen.RequestBody{
						Ref: "#/components/requestBodies/LocalRequestBody",
					},
					Responses: map[string]*ogen.Response{
						"200": {
							Ref: "#/components/responses/LocalResponse",
						},
					},
				},
			},
		},
		Components: &ogen.Components{
			Schemas: map[string]*ogen.Schema{
				"LocalSchema": {
					Ref: "foo.json#/components/schemas/RemoteSchema",
				},
			},
			Responses: map[string]*ogen.Response{
				"LocalResponse": {
					Ref: "foo.json#/components/responses/RemoteResponse",
				},
			},
			Parameters: map[string]*ogen.Parameter{
				"LocalParameter": {
					Ref: "foo.json#/components/parameters/RemoteParameter",
				},
			},
			Headers: map[string]*ogen.Header{
				"LocalHeader": {
					Ref: "foo.json#/components/headers/RemoteHeader",
				},
			},
			Examples: map[string]*ogen.Example{
				"LocalExample": {
					Ref: "foo.json#/components/examples/RemoteExample",
				},
			},
			RequestBodies: map[string]*ogen.RequestBody{
				"LocalRequestBody": {
					Ref: "foo.json#/components/requestBodies/RemoteRequestBody",
				},
			},
			SecuritySchemes: map[string]*ogen.SecuritySchema{
				"LocalSecuritySchema": {
					Ref: "foo.json#/components/securitySchemes/RemoteSecuritySchema",
				},
			},
		},
	}
	remote := external{
		"foo.json": &ogen.Spec{
			Components: &ogen.Components{
				Schemas: map[string]*ogen.Schema{
					"RemoteSchema": {
						Ref: "bar.json#/components/schemas/Schema",
					},
				},
				Responses: map[string]*ogen.Response{
					"RemoteResponse": {
						Ref: "response.json#",
					},
				},
				Parameters: map[string]*ogen.Parameter{
					"RemoteParameter": {
						Name:  "parameter",
						In:    "query",
						Style: "form",
						Schema: &ogen.Schema{
							Ref: "#/components/schemas/RemoteSchema",
						},
					},
				},
				Headers: map[string]*ogen.Header{
					"RemoteHeader": {
						Style: "simple",
						Schema: &ogen.Schema{
							Ref: "bar.json#/components/schemas/Schema",
						},
					},
				},
				Examples: map[string]*ogen.Example{
					"RemoteExample": {
						Value: exampleValue,
					},
				},
				RequestBodies: map[string]*ogen.RequestBody{
					"RemoteRequestBody": {
						Description: "request description",
						Content: map[string]ogen.Media{
							"application/json": {
								Schema: &ogen.Schema{
									Ref: "foo.json#/components/schemas/RemoteSchema",
								},
								Examples: map[string]*ogen.Example{
									"ref": {
										Ref: "#/components/examples/RemoteExample",
									},
								},
							},
						},
					},
				},
				SecuritySchemes: nil,
			},
		},
		"bar.json": &ogen.Spec{
			Components: &ogen.Components{
				Schemas: map[string]*ogen.Schema{
					"Schema": {
						Ref: "root.json#",
					},
				},
			},
		},
		"response.json": ogen.Response{
			Description: "response description",
			Headers: map[string]*ogen.Header{
				"ResponseHeader": {
					Ref: "foo.json#/components/headers/RemoteHeader",
				},
			},
			Content: map[string]ogen.Media{
				"application/json": {
					Schema: &ogen.Schema{
						Type: "string",
					},
					Examples: map[string]*ogen.Example{
						"ref": {
							Ref: "foo.json#/components/examples/RemoteExample",
						},
					},
				},
			},
		},
		"root.json": ogen.Schema{Type: "string"},
	}

	a := require.New(t)
	spec, err := Parse(root, Settings{
		External: remote,
	})
	a.NoError(err)

	var (
		schema = &jsonschema.Schema{
			Ref:      "root.json#",
			Type:     "string",
			Examples: []json.RawMessage{exampleValue},
		}
		param = &openapi.Parameter{
			Ref:     "#/components/parameters/LocalParameter",
			Name:    "parameter",
			Schema:  schema,
			In:      "query",
			Style:   "form",
			Explode: true,
		}
		requestBody = &openapi.RequestBody{
			Ref:         "#/components/requestBodies/LocalRequestBody",
			Description: "request description",
			Content: map[string]*openapi.MediaType{
				"application/json": {
					Schema: schema,
					Examples: map[string]*openapi.Example{
						"ref": {
							Ref:   "#/components/examples/RemoteExample",
							Value: exampleValue,
						},
					},
					Encoding: map[string]*openapi.Encoding{},
				},
			},
		}
		response = &openapi.Response{
			Ref:         "#/components/responses/LocalResponse",
			Description: "response description",
			Headers: map[string]*openapi.Header{
				"ResponseHeader": {
					Ref:    "foo.json#/components/headers/RemoteHeader",
					Name:   "ResponseHeader",
					In:     openapi.LocationHeader,
					Style:  openapi.HeaderStyleSimple,
					Schema: schema,
				},
			},
			Content: map[string]*openapi.MediaType{
				"application/json": {
					Schema: &jsonschema.Schema{
						Type:     "string",
						Examples: []json.RawMessage{exampleValue},
					},
					Examples: map[string]*openapi.Example{
						"ref": {
							Ref:   "foo.json#/components/examples/RemoteExample",
							Value: exampleValue,
						},
					},
					Encoding: map[string]*openapi.Encoding{},
				},
			},
		}
	)

	a.Equal(&openapi.API{
		Operations: []*openapi.Operation{
			{
				OperationID: "testGet",
				Description: "operation description",
				HTTPMethod:  "get",
				Path: openapi.Path{
					{Raw: "/get"},
				},
				Parameters:  []*openapi.Parameter{param},
				RequestBody: requestBody,
				Security:    []openapi.SecurityRequirements{},
				Responses: map[string]*openapi.Response{
					"200": response,
				},
			},
		},
		Components: &openapi.Components{
			Parameters: map[string]*openapi.Parameter{
				"LocalParameter": param,
			},
			Schemas: map[string]*jsonschema.Schema{
				"LocalSchema": schema,
			},
			RequestBodies: map[string]*openapi.RequestBody{
				"LocalRequestBody": requestBody,
			},
			Responses: map[string]*openapi.Response{
				"LocalResponse": response,
			},
		},
	}, spec)
}

func TestComplicatedReference(t *testing.T) {
	root := &ogen.Spec{
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

	raw, err := json.Marshal(root)
	a.NoError(err)
	root.Raw = raw

	spec, err := Parse(root, Settings{})
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
					Ref:  "#/paths/~1post/post/parameters/0",
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
				Ref: "#/paths/~1post/post/requestBody",
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
			Security: []openapi.SecurityRequirements{},
			Responses: map[string]*openapi.Response{
				"200": {
					Ref: "#/paths/~1post/post/responses/200",
					Headers: map[string]*openapi.Header{
						"ResponseHeader": responseHeader,
					},
					Content: map[string]*openapi.MediaType{
						"application/json": {
							Schema: &jsonschema.Schema{
								Ref:  "#/paths/~1post/post/requestBody/content/application~1json/schema",
								Type: "string",
							},
							Examples: map[string]*openapi.Example{},
							Encoding: map[string]*openapi.Encoding{},
						},
					},
				},
				"201": {
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
			Security: []openapi.SecurityRequirements{},
			Responses: map[string]*openapi.Response{
				"200": {
					Headers: map[string]*openapi.Header{
						"ResponseHeader": responseHeader,
					},
					Content: map[string]*openapi.MediaType{
						"application/json": {
							Schema: &jsonschema.Schema{
								Ref:  "#/paths/~1post/post/requestBody/content/application~1json/schema",
								Type: "string",
							},
							Examples: map[string]*openapi.Example{},
							Encoding: map[string]*openapi.Encoding{},
						},
					},
				},
			},
		}
	)
	{
		s := spec.Operations
		sort.Slice(s, func(i, j int) bool {
			return s[i].OperationID < s[j].OperationID
		})
	}
	a.Equal(&openapi.API{
		Operations: []*openapi.Operation{
			testGet,
			testPost,
		},
		Components: &openapi.Components{
			Parameters:    map[string]*openapi.Parameter{},
			Schemas:       map[string]*jsonschema.Schema{},
			RequestBodies: map[string]*openapi.RequestBody{},
			Responses:     map[string]*openapi.Response{},
		},
	}, spec)
}
