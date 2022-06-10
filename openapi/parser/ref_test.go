package parser

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
)

type external map[string]*ogen.Spec

func (e external) Get(_ context.Context, loc string) ([]byte, error) {
	r, ok := e[loc]
	if !ok {
		return nil, errors.Errorf("unexpected location %q", loc)
	}
	return json.Marshal(r)
}

func TestExternalReference(t *testing.T) {
	example := json.RawMessage(`"value"`)

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
		"foo.json": {
			Components: &ogen.Components{
				Schemas: map[string]*ogen.Schema{
					"Schema": {
						Ref: "bar.json#/components/schemas/Schema",
					},
				},
				Responses: map[string]*ogen.Response{
					"RemoteResponse": {
						Description: "response description",
						Headers: map[string]*ogen.Header{
							"ResponseHeader": {
								Ref: "#/components/headers/RemoteHeader",
							},
						},
						Content: map[string]ogen.Media{
							"application/json": {
								Schema: &ogen.Schema{
									Type: "string",
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
				Parameters: map[string]*ogen.Parameter{
					"RemoteParameter": {
						Name:  "parameter",
						In:    "query",
						Style: "form",
						Schema: &ogen.Schema{
							Ref: "#/components/schemas/Schema",
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
						Value: example,
					},
				},
				RequestBodies: map[string]*ogen.RequestBody{
					"RemoteRequestBody": {
						Description: "request description",
						Content: map[string]ogen.Media{
							"application/json": {
								Schema: &ogen.Schema{
									Ref: "foo.json#/components/schemas/Schema",
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
		"bar.json": {
			Components: &ogen.Components{
				Schemas: map[string]*ogen.Schema{
					"Schema": {
						Type: "string",
					},
				},
			},
		},
	}

	a := require.New(t)
	spec, err := Parse(root, Settings{
		External: remote,
	})
	a.NoError(err)

	var (
		schema = &jsonschema.Schema{
			Ref:      "bar.json#/components/schemas/Schema",
			Type:     "string",
			Examples: []json.RawMessage{example},
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
							Value: example,
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
					Ref:    "#/components/headers/RemoteHeader",
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
						Examples: []json.RawMessage{example},
					},
					Examples: map[string]*openapi.Example{
						"ref": {
							Ref:   "#/components/examples/RemoteExample",
							Value: example,
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
			Schemas: map[string]*jsonschema.Schema{},
			RequestBodies: map[string]*openapi.RequestBody{
				"LocalRequestBody": requestBody,
			},
			Responses: map[string]*openapi.Response{
				"LocalResponse": response,
			},
		},
	}, spec)
}
