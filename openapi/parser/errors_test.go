package parser

import (
	"encoding/json"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/location"
)

func TestRemoteLocation(t *testing.T) {
	a := require.New(t)

	root := &ogen.Spec{
		OpenAPI: "3.0.3",
		Paths: map[string]*ogen.PathItem{
			"/get": {
				Get: &ogen.Operation{
					OperationID: "testGet",
					Description: "operation description",
					Responses: map[string]*ogen.Response{
						"200": {},
					},
				},
			},
		},
		Components: &ogen.Components{
			Parameters: map[string]*ogen.Parameter{
				"LocalParameter": {
					Ref: "foo.json#/components/parameters/RemoteParameter",
				},
			},
		},
	}
	remote := external{
		"foo.json": json.RawMessage(`{"components": {"parameters": {"RemoteParameter": {"in": "amongus"}}}}`),
	}

	_, err := Parse(root, Settings{
		External: remote,
		File:     location.NewFile("root.json", "root.json", nil),
	})
	a.Error(err)
	var (
		iterErr = err
		locErr  *LocationError
	)
	for {
		if !errors.As(iterErr, &locErr) {
			break
		}
		iterErr = locErr.Err
	}
	t.Log(locErr)
	a.NotNil(locErr)
	a.Equal("foo.json", locErr.File.Name)
}
