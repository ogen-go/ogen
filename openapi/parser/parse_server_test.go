package parser_test

import (
	"testing"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/openapi/parser"
	"github.com/stretchr/testify/require"
)

func TestServerURL(t *testing.T) {
	data := `
info:
  description: test
  title: test
  version: 1.0.0
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
`
	spec, err := ogen.Parse([]byte(data))
	require.NoError(t, err)

	api, err := parser.Parse(spec, parser.Settings{})
	require.NoError(t, err)

	expandSpec, err := parser.Expand(api)
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
}
