package parser

import (
	"encoding/json"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
)

func TestRemoteLocation(t *testing.T) {
	a := require.New(t)

	root := &ogen.Spec{
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
		Filename: "root.json",
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
	loc := locErr.Loc
	a.Equal("foo.json", loc.Filename)
}
