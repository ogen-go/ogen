package gen

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/jsonschema"
)

func TestParseSimpleIntegerFormats(t *testing.T) {
	tests := []struct {
		format string
		want   ir.PrimitiveType
	}{
		{format: "", want: ir.Int},
		{format: "unknown", want: ir.Int},
		{format: "int64", want: ir.Int64},
		{format: "unixtime", want: ir.Time},
		{format: "unix", want: ir.Time},
		{format: "unix-seconds", want: ir.Time},
		{format: "unix-nano", want: ir.Time},
		{format: "unix-micro", want: ir.Time},
		{format: "unix-milli", want: ir.Time},
	}
	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			schema := &jsonschema.Schema{
				Type:   jsonschema.Integer,
				Format: tt.format,
			}
			g := newSchemaGen(func(jsonschema.Ref) (*ir.Type, bool) { return nil, false })
			typ := g.parseSimple(schema)
			require.Equal(t, ir.KindPrimitive, typ.Kind)
			require.Equal(t, tt.want, typ.Primitive)
			require.Same(t, schema, typ.Schema)
		})
	}
}

func TestGenerateUnixTime(t *testing.T) {
	const input = `openapi: 3.0.0
info:
  title: Unix timestamp
  version: 1.0.0
paths:
  /timestamp:
    get:
      operationId: getTimestamp
      parameters:
        - name: since
          in: query
          schema:
            type: %s
            format: unixtime
      responses:
        '200':
          description: Unix timestamp in seconds
          content:
            application/json:
              schema:
                type: object
                required: [timestamp]
                properties:
                  timestamp:
                    type: %s
                    format: unixtime
                  optionalTimestamp:
                    type: %s
                    format: unixtime
`
	for _, tt := range []struct {
		schemaType string
		formatName string
	}{
		{schemaType: "integer", formatName: "UnixSeconds"},
		{schemaType: "string", formatName: "StringUnixSeconds"},
	} {
		t.Run(tt.schemaType, func(t *testing.T) {
			spec, err := ogen.Parse(fmt.Appendf(nil, input, tt.schemaType, tt.schemaType, tt.schemaType))
			require.NoError(t, err)

			_, fs := generateSpecTest(t, spec, Options{})
			require.Regexp(t, `Timestamp\s+time\.Time`, fs.Get("oas_schemas_gen.go"))
			require.Contains(t, fs.Get("oas_schemas_gen.go"), "type Opt"+tt.formatName+" struct")
			require.Contains(t, fs.Get("oas_json_gen.go"), "json.Decode"+tt.formatName+"(d)")
			require.Contains(t, fs.Get("oas_json_gen.go"), "json.Encode"+tt.formatName+"(e, s.Timestamp)")
			require.Contains(t, fs.Get("oas_json_gen.go"), "s.OptionalTimestamp.Encode(e, json.Encode"+tt.formatName+")")
			require.Contains(t, fs.Get("oas_client_gen.go"), "conv.UnixSecondsToString(")
			require.Contains(t, fs.Get("oas_parameters_gen.go"), "conv.ToUnixSeconds(")
		})
	}
}
