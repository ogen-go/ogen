package gen

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
)

func TestGenerateHeadersReusesEquivalentHeaders(t *testing.T) {
	g := &Generator{
		log: zap.NewNop(),
		parseOpts: ParseOptions{
			SchemaDepthLimit: defaultSchemaDepthLimit,
		},
	}
	global := newTStorage()
	firstCtx := &genctx{global: global, local: newTStorage()}
	secondCtx := &genctx{global: global, local: newTStorage()}

	first, err := g.generateHeaders(firstCtx, "FirstGetTooManyRequests", map[string]*openapi.Header{
		"Retry-After": {Name: "Retry-After", Schema: retryAfterSchema()},
	})
	require.NoError(t, err)
	require.NoError(t, global.merge(firstCtx.local))

	second, err := g.generateHeaders(secondCtx, "SecondGetTooManyRequests", map[string]*openapi.Header{
		"Retry-After": {Name: "Retry-After", Schema: retryAfterSchema()},
	})
	require.NoError(t, err)

	require.Same(t, first["Retry-After"], second["Retry-After"])
}

func TestGenerateHeadersDoesNotReuseDifferentHeaders(t *testing.T) {
	g := &Generator{
		log: zap.NewNop(),
		parseOpts: ParseOptions{
			SchemaDepthLimit: defaultSchemaDepthLimit,
		},
	}
	global := newTStorage()
	firstCtx := &genctx{global: global, local: newTStorage()}
	secondCtx := &genctx{global: global, local: newTStorage()}

	first, err := g.generateHeaders(firstCtx, "FirstGet", map[string]*openapi.Header{
		"Retry-After": {Name: "Retry-After", Schema: &jsonschema.Schema{Type: jsonschema.Integer}},
	})
	require.NoError(t, err)
	require.NoError(t, global.merge(firstCtx.local))

	second, err := g.generateHeaders(secondCtx, "SecondGet", map[string]*openapi.Header{
		"Retry-After": {Name: "Retry-After", Schema: &jsonschema.Schema{Type: jsonschema.String}},
	})
	require.NoError(t, err)

	require.NotSame(t, first["Retry-After"], second["Retry-After"])
}

func retryAfterSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		AnyOf: []*jsonschema.Schema{
			{Type: jsonschema.Integer},
			{Type: jsonschema.String},
		},
	}
}
