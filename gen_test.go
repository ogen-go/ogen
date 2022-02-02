package ogen_test

import (
	"embed"
	"go/format"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen"
)

//go:embed _testdata
var testdata embed.FS

// TODO: Create validationFs.
type fmtFs struct{}

func (n fmtFs) WriteFile(baseName string, source []byte) error {
	_, err := format.Source(source)
	return err
}

func testGenerate(t *testing.T, name string, ignore ...string) {
	t.Helper()

	data, err := testdata.ReadFile(path.Join("_testdata", name))
	require.NoError(t, err)
	spec, err := ogen.Parse(data)
	require.NoError(t, err)
	opt := gen.Options{
		IgnoreNotImplemented: ignore,
		InferSchemaType:      true,
	}
	t.Run("Gen", func(t *testing.T) {
		g, err := gen.NewGenerator(spec, opt)
		require.NoError(t, err)

		require.NoError(t, g.WriteSource(fmtFs{}, "api"))
	})
	if len(opt.IgnoreNotImplemented) > 0 {
		t.Run("Full", func(t *testing.T) {
			t.Skipf("Ignoring: %s", opt.IgnoreNotImplemented)
		})
	}
}

func TestGenerate(t *testing.T) {
	t.Parallel()
	g := func(name string, ignore ...string) func(t *testing.T) {
		return func(t *testing.T) {
			t.Helper()
			t.Parallel()
			testGenerate(t, name, ignore...)
		}
	}

	t.Run("Pet store", g("petstore.yaml"))
	t.Run("Pet store expanded", g("petstore-expanded.yaml",
		"allOf",
	))
	t.Run("Firecracker", g("firecracker.json"))
	t.Run("Sample", g("sample.json"))
	t.Run("Manga gallery", g("manga.json",
		"unsupported content types",
	))
	t.Run("TechEmpower", g("techempower.json"))
	t.Run("telegram bot api", g("telegram_bot_api.json",
		"anyOf",
		"unsupported content types",
	))
	t.Run("gotd botapi", g("gotd_bot_api.json",
		"unsupported content types",
	))
	t.Run("Kubernetes", g("k8s.json",
		"unsupported content types",
	))
	t.Run("GitHub", g("api.github.com.json",
		"complex parameter types",
		"anyOf",
		"allOf",
		"discriminator inference",
		"sum types with same names",
		"sum type parameter",
		"unsupported content types",
		"empty schema",
	))
	t.Run("Tinkoff", g("tinkoff.json"))
}
