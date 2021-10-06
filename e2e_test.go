package ogen_test

import (
	"embed"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/gen"
)

//go:embed _testdata
var testdata embed.FS

type noopFileSystem struct{}

func (n noopFileSystem) WriteFile(baseName string, source []byte) error { return nil }

func TestGenerate(t *testing.T) {
	for _, tc := range []struct {
		Name    string
		Options []gen.Option
	}{
		{
			Name: "firecracker.json",
			Options: []gen.Option{
				gen.WithIgnoreOptionals,
				gen.WithIgnoreFormat,
				gen.WithEmptyObjects,
			},
		},
		{
			Name: "api.github.com.json",
			Options: []gen.Option{
				gen.WithIgnoreOptionals,
				gen.WithIgnoreFormat,
				gen.WithEmptyObjects,
				gen.WithSkipUnspecified,
			},
		},
		{
			Name: "sample_1.json",
		},
		{
			Name: "techempower.json",
		},
		{
			Name: "telegram_bot_api.json",
			Options: []gen.Option{
				gen.WithIgnoreOptionals,
				gen.WithIgnoreFormat,
				gen.WithSkipUnspecified,
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			f, err := testdata.Open(path.Join("_testdata", tc.Name))
			require.NoError(t, err)
			defer require.NoError(t, f.Close())
			spec, err := ogen.Parse(f)
			require.NoError(t, err)
			g, err := gen.NewGenerator(spec, tc.Options...)
			require.NoError(t, err)

			require.NoError(t, g.WriteSource(noopFileSystem{}, "api"))
		})
	}
}
