package parser

import (
	"context"
	"embed"
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strings"
	"testing"

	"github.com/go-faster/errors"
	yaml "github.com/go-faster/yamlx"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/location"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
)

type external map[string]any

func (e external) Get(_ context.Context, loc string) ([]byte, error) {
	loc = strings.TrimPrefix(loc, "/")
	r, ok := e[loc]
	if !ok {
		return nil, errors.Errorf("unexpected location %q", loc)
	}
	return json.Marshal(r)
}

func TestExternalReference(t *testing.T) {
	exampleValue := jsonschema.Example(`"value"`)

	root := &ogen.Spec{
		OpenAPI: "3.1.0",
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
			"/pathItem": {
				Ref: "#/components/pathItems/LocalPathItem",
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
			Headers: map[string]*ogen.Header{
				"LocalHeader": {
					Ref: "foo.json#/components/headers/RemoteHeader",
				},
			},
			SecuritySchemes: map[string]*ogen.SecurityScheme{
				"LocalSecurityScheme": {
					Ref: "foo.json#/components/securitySchemes/RemoteSecurityScheme",
				},
			},
			PathItems: map[string]*ogen.PathItem{
				"LocalPathItem": {
					Ref: "pathItem.json",
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
		"pathItem.json": &ogen.PathItem{
			Get: &ogen.Operation{
				OperationID: "remoteGet",
				Description: "remote operation description",
				Responses: map[string]*ogen.Response{
					"200": {
						Ref: "response.json#",
					},
				},
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
			Examples: []jsonschema.Example{exampleValue},
		}
		localExample = &openapi.Example{
			Ref:   "foo.json#/components/examples/RemoteExample",
			Value: exampleValue,
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
							Ref:   "foo.json#/components/examples/RemoteExample",
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
						Examples: []jsonschema.Example{exampleValue},
					},
					Examples: map[string]*openapi.Example{
						"ref": localExample,
					},
					Encoding: map[string]*openapi.Encoding{},
				},
			},
		}
	)

	compareJSON(t, &openapi.API{
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
				Security:    []openapi.SecurityRequirement{},
				Responses: map[string]*openapi.Response{
					"200": response,
				},
			},
			{
				OperationID: "remoteGet",
				Description: "remote operation description",
				HTTPMethod:  "get",
				Path: openapi.Path{
					{Raw: "/pathItem"},
				},
				Parameters:  []*openapi.Parameter{},
				RequestBody: nil,
				Security:    []openapi.SecurityRequirement{},
				Responses: map[string]*openapi.Response{
					"200": response,
				},
			},
		},
		Components: &openapi.Components{
			Schemas: map[string]*jsonschema.Schema{
				"LocalSchema": schema,
			},
			Responses: map[string]*openapi.Response{
				"LocalResponse": response,
			},
			Parameters: map[string]*openapi.Parameter{
				"LocalParameter": param,
			},
			Examples: map[string]*openapi.Example{
				"LocalExample": {
					Ref:   "#/components/examples/LocalExample",
					Value: exampleValue,
				},
			},
			RequestBodies: map[string]*openapi.RequestBody{
				"LocalRequestBody": requestBody,
			},
		},
	}, spec)
}

// Ensure that parser checks for duplicate operation IDs even if there is pathItem reference.
func TestDuplicateOperationID(t *testing.T) {
	root := &ogen.Spec{
		OpenAPI: "3.0.3",
		Paths: map[string]*ogen.PathItem{
			"/get": {
				Get: &ogen.Operation{
					OperationID: "testGet",
					Description: "local",
					Responses: map[string]*ogen.Response{
						"200": {
							Description: "response description",
						},
					},
				},
			},
			"/pathItem": {
				Ref: "pathItem.json#",
			},
		},
		Components: &ogen.Components{},
	}
	remote := external{
		"pathItem.json": &ogen.PathItem{
			Get: &ogen.Operation{
				OperationID: "testGet",
				Description: "remote",
				Responses: map[string]*ogen.Response{
					"200": {
						Description: "response description",
					},
				},
			},
		},
	}

	a := require.New(t)
	_, err := Parse(root, Settings{
		External: remote,
	})
	a.ErrorContains(err, "duplicate operationId: \"testGet\"")
	// Ensure that the error contains the file name.
	var locErr *location.Error
	a.ErrorAs(err, &locErr)
	a.Equal("pathItem.json", locErr.File.Name)
}

// Ensure that parser adds location information to the error, even if the error is occurred in the external file.
func TestExternalErrors(t *testing.T) {
	root := &ogen.Spec{
		OpenAPI: "3.0.3",
		Paths: map[string]*ogen.PathItem{
			"/pathItem": {
				Ref: "pathItem.json#",
			},
		},
		Components: &ogen.Components{},
	}
	remote := external{
		"pathItem.json": &ogen.PathItem{
			Get: &ogen.Operation{
				Responses: map[string]*ogen.Response{
					"very bad status code": {},
				},
			},
		},
	}

	a := require.New(t)
	_, err := Parse(root, Settings{
		External: remote,
	})
	a.ErrorContains(err, "parse status code")

	var locErr *location.Error
	a.ErrorAs(err, &locErr)
	a.Equal("pathItem.json", locErr.File.Name)
	a.True(locErr.PrettyPrint(io.Discard, false))
}

//go:embed _testdata/remotes
var remotes embed.FS

// TestInitialLocation ensures that the parser respects the RootURL setting.
func TestInitialLocation(t *testing.T) {
	const (
		rootSpec   = "api/spec.json"
		rootPrefix = "_testdata/remotes"
	)

	raw, spec := func() ([]byte, *ogen.Spec) {
		raw, err := fs.ReadFile(remotes, path.Join(rootPrefix, rootSpec))
		require.NoError(t, err)

		var spec ogen.Spec
		require.NoError(t, yaml.Unmarshal(raw, &spec))
		return raw, &spec
	}()

	check := func(a *require.Assertions, parsed *openapi.API) {
		ops := parsed.Operations
		a.Len(ops, 1)
		op := ops[0]
		a.Equal("/get", op.Path.String())

		resp, ok := op.Responses["200"]
		a.True(ok)
		content, ok := resp.Content["application/json"]
		a.True(ok)

		s := content.Schema
		a.Equal(jsonschema.Object, s.Type)
		a.Len(s.Properties, 2)
		propAge := s.Properties[1]
		a.Equal("age", propAge.Name)
		a.Equal(jsonschema.Integer, propAge.Schema.Type)
	}

	t.Run("HTTP", func(t *testing.T) {
		a := require.New(t)

		remotes, err := fs.Sub(remotes, rootPrefix)
		a.NoError(err)

		h := http.FileServer(http.FS(remotes))
		srv := httptest.NewServer(h)
		defer srv.Close()

		rootURL, err := url.Parse(srv.URL)
		a.NoError(err)
		rootURL.Path = path.Join(rootURL.Path, rootSpec)

		parsed, err := Parse(spec, Settings{
			External: jsonschema.NewExternalResolver(jsonschema.ExternalOptions{
				HTTPClient: srv.Client(),
				ReadFile: func(p string) ([]byte, error) {
					return nil, errors.Errorf("unexpected call: %q", p)
				},
			}),
			File:    location.NewFile(rootSpec, rootURL.String(), raw),
			RootURL: rootURL,
		})
		a.NoError(err)
		check(a, parsed)
	})
}
