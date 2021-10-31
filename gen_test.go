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
	t.Parallel()

	for _, tc := range []struct {
		Name    string
		Options gen.Options
	}{
		{
			Name: "petstore-expanded.yaml",
		},
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
			Name: "sample.json",
		},
		{
			Name: "nh.json",
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
			Name: "gotd_bot_api.json",
		},
		{
			// https://github.com/kubernetes/kubernetes/tree/master/api/openapi-spec
			// Generated from OpenAPI v2 (swagger) spec.
			Name: "k8s.json",
			Options: gen.Options{
				IgnoreUnspecifiedParams: true,
				IgnoreNotImplemented: []string{
					"requestBody with primitive type",
					"response with primitive type",
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			data, err := testdata.ReadFile(path.Join("_testdata", tc.Name))
			require.NoError(t, err)
			spec, err := ogen.Parse(data)
			require.NoError(t, err)
			g, err := gen.NewGenerator(spec, tc.Options)
			require.NoError(t, err)

			require.NoError(t, g.WriteSource(fmtFs{}, "api"))
		})
	}
}
