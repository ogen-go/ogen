package ogen_test

import (
	"embed"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen"
	"github.com/ogen-go/ogen/gen/genfs"
)

//go:embed _testdata
var testdata embed.FS

func testGenerate(t *testing.T, _ bool, name string, ignore ...string) {
	t.Helper()

	data, err := testdata.ReadFile(name)
	require.NoError(t, err)
	spec, err := ogen.Parse(data)
	require.NoError(t, err)
	opt := gen.Options{
		IgnoreNotImplemented: ignore,
		InferSchemaType:      true,
	}
	t.Run("Gen", func(t *testing.T) {
		defer func() {
			if rr := recover(); rr != nil {
				t.Fatalf("panic: %+v", rr)
			}
		}()

		g, err := gen.NewGenerator(spec, opt)
		require.NoError(t, err)
		require.NoError(t, g.WriteSource(genfs.CheckFS{}, "api"))
	})
	if len(opt.IgnoreNotImplemented) > 0 {
		t.Run("Full", func(t *testing.T) {
			t.Skipf("Ignoring: %s", opt.IgnoreNotImplemented)
		})
	}
}

func TestGenerate(t *testing.T) {
	t.Parallel()
	g := func(name string, build bool, ignore ...string) func(t *testing.T) {
		return func(t *testing.T) {
			t.Helper()
			t.Parallel()
			testGenerate(t, build, name, ignore...)
		}
	}

	skipSets := map[string][]string{
		"petstore.yaml": {},
		"petstore-expanded.yaml": {
			"allOf",
		},
		"firecracker.json": {},
		"sample.json":      {},
		"manga.json": {
			"unsupported content types",
		},
		"techempower.json": {},
		"telegram_bot_api.json": {
			"anyOf",
			"unsupported content types",
		},
		"gotd_bot_api.json": {
			"unsupported content types",
		},
		"k8s.json": {
			"unsupported content types",
		},
		"api.github.com.json": {
			"complex parameter types",
			"complex anyOf",
			"allOf",
			"discriminator inference",
			"sum types with same names",
			"sum type parameter",
			"unsupported content types",
			"empty schema",
		},
		"test_empty_property_name.yaml": {},
		"tinkoff.json": {
			"http security scheme",
		},
	}
	// Do not build these schemas because they checked by test examples workflow.
	doNotBuild := map[string]struct{}{
		"test_format.json":         {},
		"petstore.yaml":            {},
		"firecracker.json":         {},
		"gotd_bot_api.json":        {},
		"ent.json":                 {},
		"ex_route_params.json":     {},
		"manga.json":               {},
		"petstore - expanded.yaml": {},
		"k8s.json":                 {},
		"api.github.com.json":      {},
		"telegram_bot_api.json":    {},
		"tinkoff.json":             {},
	}

	testDataPath := "_testdata/positive"
	if err := fs.WalkDir(testdata, testDataPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return err
		}

		_, file := filepath.Split(path)
		skip, ok := skipSets[file]
		if !ok {
			skip = []string{"all"}
		}

		build := !testing.Short()
		if _, ok := doNotBuild[file]; ok {
			build = false
		}

		testName := strings.TrimPrefix(path, testDataPath+"/")
		testName = strings.TrimSuffix(testName, ".json")
		testName = strings.TrimSuffix(testName, ".yml")
		testName = strings.TrimSuffix(testName, ".yaml")

		t.Run(testName, g(path, build, skip...))
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
