package ogen_test

import (
	"embed"
	"go/format"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/gen"
)

//go:embed _testdata
var testdata embed.FS

// TODO: Create validationFs.
type fmtFs struct{}

func (n fmtFs) WriteFile(baseName string, source []byte) error {
	_, err := format.Source(source)
	return err
}

func TestGenerate(t *testing.T) {
	for _, tc := range []struct {
		Name    string
		Options gen.Options
	}{
		{
			Name: "firecracker.json",
		},
		{
			Name: "api.github.com.json",
			Options: gen.Options{
				IgnoreNotImplemented: []string{
					"complex parameter types",
					"oneOf",
					"anyOf",
					"allOf",
					"nullable",
					"array parameter with complex type",
					"optional nullable array",
				},
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
			Options: gen.Options{
				IgnoreNotImplemented: []string{"anyOf"},
			},
		},
		{
			// https://github.com/kubernetes/kubernetes/tree/master/api/openapi-spec
			// Generated from OpenAPI v2 (swagger) spec.
			Name: "k8s.json",
			Options: gen.Options{
				IgnoreUnspecifiedParams: true,
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			f, err := testdata.Open(path.Join("_testdata", tc.Name))
			require.NoError(t, err)
			defer require.NoError(t, f.Close())
			spec, err := ogen.Parse(f)
			require.NoError(t, err)
			g, err := gen.NewGenerator(spec, tc.Options)
			require.NoError(t, err)

			require.NoError(t, g.WriteSource(fmtFs{}, "api"))
		})
	}
}
