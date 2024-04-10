package parser

import (
	"testing"

	"github.com/go-faster/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
)

// Check that parser correctly handles paths with escaped slashes.
func TestEscapedSlashInPath(t *testing.T) {
	// If you unescape "/user%2Fget", it will be "/user/get".
	//
	// But this slash is not the path separator, it's a part of the path.
	//
	// Parser should keep it as is, since these paths are two different paths.
	root := &ogen.Spec{
		OpenAPI: "3.0.3",
		Paths: map[string]*ogen.PathItem{
			"/user/get": {
				Get: &ogen.Operation{
					OperationID: "userGet",
					Description: "operation description",
					Responses: map[string]*ogen.Response{
						"200": {},
					},
				},
			},
			"/user%2Fget": {
				Get: &ogen.Operation{
					OperationID: "escapedUserGet",
					Description: "operation description",
					Responses: map[string]*ogen.Response{
						"200": {},
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
	a.Len(spec.Operations, 2)
}

func TestXOgenOperationGroup(t *testing.T) {
	root := &ogen.Spec{
		OpenAPI: "3.0.3",
		Paths: map[string]*ogen.PathItem{
			"/users": {
				Common: extensionValue(xOgenOperationGroup, "Users"),
				Get: &ogen.Operation{
					OperationID: "userList",
					Responses: map[string]*ogen.Response{
						"200": {},
					},
				},
				Post: &ogen.Operation{
					Common:      extensionValue(xOgenOperationGroup, "Override"),
					OperationID: "userCreate",
					Responses: map[string]*ogen.Response{
						"200": {},
					},
				},
			},
		},
	}

	spec, err := Parse(root, Settings{
		RootURL: testRootURL,
	})
	assert.NoError(t, err)

	expected := &openapi.API{
		Version: openapi.Version{Major: 3, Minor: 0, Patch: 3},
		Operations: []*openapi.Operation{
			{
				OperationID: "userList",
				HTTPMethod:  "get",
				Path: openapi.Path{
					{Raw: "/users"},
				},
				Parameters: []*openapi.Parameter{},
				Security:   openapi.SecurityRequirements{},
				Responses: openapi.Responses{
					StatusCode: map[int]*openapi.Response{
						200: {},
					},
				},
				XOgenOperationGroup: "Users",
			},
			{
				OperationID: "userCreate",
				HTTPMethod:  "post",
				Path: openapi.Path{
					{Raw: "/users"},
				},
				Parameters: []*openapi.Parameter{},
				Security:   openapi.SecurityRequirements{},
				Responses: openapi.Responses{
					StatusCode: map[int]*openapi.Response{
						200: {},
					},
				},
				XOgenOperationGroup: "Override",
			},
		},
		Components: &openapi.Components{
			Schemas:       map[string]*jsonschema.Schema{},
			Responses:     map[string]*openapi.Response{},
			Parameters:    map[string]*openapi.Parameter{},
			Examples:      map[string]*openapi.Example{},
			RequestBodies: map[string]*openapi.RequestBody{},
		},
	}

	assert.Equal(t, expected, spec)
}

func extensionValue(name, value string) ogen.OpenAPICommon {
	return ogen.OpenAPICommon{
		Extensions: ogen.Extensions{
			name: yaml.Node{Kind: yaml.ScalarNode, Value: value},
		},
	}
}
