package parser_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/openapi/parser"
)

func TestServerURL(t *testing.T) {
	data := `
info:
  description: test
  title: test
  version: 1.0.0
tags:
  - name: demo1
    description: demo1 description
  - name: demo2
    description: demo2 description
openapi: "3.0.0"
servers:
  - url: "{protocol}://{host}:{port}"
    variables:
      host:
        default: localhost
      port:
        default: "4000"
      protocol:
        default: http
        enum:
          - http
          - https
paths:
  /signup:
    post:
      tags:
        - demo1
      operationId: signup
      responses:
        '204':
          description: Successful operation
        'default':
              Error:
      description: Some error during request processing
      content:
        application/json:
          schema:
            type: object
            required:
              - error
            properties:
              error:
                type: string'
`
	spec, err := ogen.Parse([]byte(data))
	require.NoError(t, err)

	api, err := parser.Parse(spec, parser.Settings{})
	require.NoError(t, err)

	expandSpec, err := parser.Expand(api, spec)
	require.NoError(t, err)

	require.Equal(t, []ogen.Server{
		{
			URL: "{protocol}://{host}:{port}",
			Variables: map[string]ogen.ServerVariable{
				"host": {
					Default: "localhost",
				},
				"protocol": {
					Enum:    []string{"http", "https"},
					Default: "http",
				},
				"port": {
					Default: "4000",
				},
			},
		},
	}, expandSpec.Servers)
	require.Equal(t, "demo1", expandSpec.Tags[0].Name)
	require.Equal(t, "demo2", expandSpec.Tags[1].Name)
	require.Equal(t, "demo1 description", expandSpec.Tags[0].Description)
	require.Equal(t, "demo2 description", expandSpec.Tags[1].Description)
	require.Equal(t, []string{"demo1"}, expandSpec.Paths["/signup"].Post.Tags)
}
