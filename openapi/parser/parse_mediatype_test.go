package parser_test

import (
	"testing"

	"github.com/go-faster/yaml"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/openapi"
	"github.com/ogen-go/ogen/openapi/parser"
)

func extensionString(v string) yaml.Node {
	return yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: v}
}

func extensionBool(v bool) yaml.Node {
	if v {
		return yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "true"}
	}
	return yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "false"}
}

func sseSpec(media ogen.Media) *ogen.Spec {
	return &ogen.Spec{
		OpenAPI: "3.0.0",
		Info: ogen.Info{
			Title:   "test",
			Version: "0.0.0",
		},
		Paths: map[string]*ogen.PathItem{
			"/stream": {
				Get: &ogen.Operation{
					OperationID: "stream",
					Responses: map[string]*ogen.Response{
						"200": {
							Content: map[string]ogen.Media{
								"text/event-stream": media,
							},
						},
					},
				},
			},
		},
	}
}

func TestParseMediaTypeSSEShapeDefault(t *testing.T) {
	api, err := parser.Parse(sseSpec(ogen.Media{
		Schema: &ogen.Schema{Type: "object"},
	}), parser.Settings{})
	require.NoError(t, err)

	media := api.Operations[0].Responses.StatusCode[200].Content["text/event-stream"]
	require.Equal(t, openapi.SSEEventShapeDataOnly, media.XOgenSSEEventShape)
}

func TestParseMediaTypeSSEShapeFull(t *testing.T) {
	api, err := parser.Parse(sseSpec(ogen.Media{
		Schema: &ogen.Schema{Type: "object"},
		Common: ogen.OpenAPICommon{
			Extensions: ogen.Extensions{
				"x-ogen-sse-event-shape": extensionString("full"),
			},
		},
	}), parser.Settings{})
	require.NoError(t, err)

	media := api.Operations[0].Responses.StatusCode[200].Content["text/event-stream"]
	require.Equal(t, openapi.SSEEventShapeFull, media.XOgenSSEEventShape)
}

func TestParseMediaTypeSSEShapeFullArray(t *testing.T) {
	api, err := parser.Parse(sseSpec(ogen.Media{
		Schema: &ogen.Schema{
			Type: "array",
			Items: &ogen.Items{
				Item: &ogen.Schema{
					Type: "object",
				},
			},
		},
		Common: ogen.OpenAPICommon{
			Extensions: ogen.Extensions{
				"x-ogen-sse-event-shape": extensionString("full-array"),
			},
		},
	}), parser.Settings{})
	require.NoError(t, err)

	media := api.Operations[0].Responses.StatusCode[200].Content["text/event-stream"]
	require.Equal(t, openapi.SSEEventShapeFullArray, media.XOgenSSEEventShape)
}

func TestParseMediaTypeSSEShapeRawResponsePriority(t *testing.T) {
	api, err := parser.Parse(sseSpec(ogen.Media{
		Schema: &ogen.Schema{Type: "object"},
		Common: ogen.OpenAPICommon{
			Extensions: ogen.Extensions{
				"x-ogen-sse-event-shape": extensionString("full"),
				"x-ogen-raw-response":    extensionBool(true),
			},
		},
	}), parser.Settings{})
	require.NoError(t, err)

	media := api.Operations[0].Responses.StatusCode[200].Content["text/event-stream"]
	require.Equal(t, openapi.SSEEventShapeNone, media.XOgenSSEEventShape)
	require.True(t, media.XOgenRawResponse)
}

func TestParseMediaTypeSSEShapeNonEventStreamError(t *testing.T) {
	spec := sseSpec(ogen.Media{})
	spec.Paths["/stream"].Get.Responses["200"].Content["application/json"] = ogen.Media{
		Schema: &ogen.Schema{Type: "object"},
		Common: ogen.OpenAPICommon{
			Extensions: ogen.Extensions{
				"x-ogen-sse-event-shape": extensionString("data-only"),
			},
		},
	}
	delete(spec.Paths["/stream"].Get.Responses["200"].Content, "text/event-stream")

	_, err := parser.Parse(spec, parser.Settings{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "x-ogen-sse-event-shape")
	require.Contains(t, err.Error(), "is only allowed for text/event-stream media type")
}
