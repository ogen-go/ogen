package gen

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
)

func TestGenerator_reduceDefault_JSONContentTypes(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		aliases     ContentTypeAliases
		want        ir.Encoding
		wantErr     bool
	}{
		{
			name:        "problem-json",
			contentType: "application/problem+json",
			want:        ir.EncodingProblemJSON,
		},
		{
			name:        "vendor-json-suffix",
			contentType: "application/vnd.api+json",
			want:        ir.EncodingJSON,
		},
		{
			name:        "non-json",
			contentType: "text/plain",
			wantErr:     true,
		},
		{
			name:        "json-suffix-aliased-to-text",
			contentType: "application/vnd.api+json",
			aliases: ContentTypeAliases{
				"application/vnd.api+json": ir.EncodingTextPlain,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Generator{
				opt: GenerateOptions{
					ConvenientErrors:   1,
					ContentTypeAliases: tt.aliases,
				},
				parseOpts: ParseOptions{SchemaDepthLimit: defaultSchemaDepthLimit},
				tstorage:  newTStorage(),
				log:       zap.NewNop(),
			}

			op := &openapi.Operation{
				Responses: openapi.Responses{
					Default: &openapi.Response{
						Content: map[string]*openapi.MediaType{
							tt.contentType: {
								Schema: &jsonschema.Schema{Type: jsonschema.Object},
							},
						},
					},
				},
			}

			a := require.New(t)
			err := g.reduceDefault([]*openapi.Operation{op})
			if tt.wantErr {
				a.Error(err)
				return
			}

			a.NoError(err)
			a.NotNil(g.errType)
			a.Len(g.errType.Contents, 1)

			for _, media := range g.errType.Contents {
				a.Equal(tt.want, media.Encoding)
			}
		})
	}
}
