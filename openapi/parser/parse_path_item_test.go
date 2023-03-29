package parser

import (
	"testing"

	"github.com/go-faster/yaml"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
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
