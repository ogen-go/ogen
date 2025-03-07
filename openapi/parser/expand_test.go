package parser_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/openapi/parser"
)

func TestExpand(t *testing.T) {
	f, err := os.ReadFile("_testdata/expand/info_tags.yaml")
	require.NoError(t, err)

	spec, err := ogen.Parse(f)
	require.NoError(t, err)
	require.NotEmpty(t, spec.Tags)
	require.Len(t, spec.Tags, 3)

	api, err := parser.Parse(spec, parser.Settings{})
	require.NoError(t, err)

	expandSpec, err := parser.Expand(api)
	require.NoError(t, err)

	require.Equal(t, "info_test", expandSpec.Info.Title)
	require.Equal(t, "1.0.0", expandSpec.Info.Version)
	require.Empty(t, expandSpec.Info.Summary)
	require.Len(t, expandSpec.Tags, 3)
}
