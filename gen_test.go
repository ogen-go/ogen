package ogen_test

import (
	"embed"
	"io/fs"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen"
	"github.com/ogen-go/ogen/gen/genfs"
)

//go:embed _testdata
var testdata embed.FS

func testGenerate(name string, ignore ...string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()
		t.Parallel()
		log := zaptest.NewLogger(t)

		data, err := testdata.ReadFile(name)
		require.NoError(t, err)
		spec, err := ogen.Parse(data)
		require.NoError(t, err)

		notImplemented := map[string]struct{}{}
		opt := gen.Options{
			InferSchemaType:      true,
			IgnoreNotImplemented: ignore,
			NotImplementedHook: func(name string, err error) {
				notImplemented[name] = struct{}{}
			},
			Logger: log,
		}
		t.Run("Gen", func(t *testing.T) {
			defer func() {
				if rr := recover(); rr != nil {
					t.Fatalf("panic: %+v\n%s", rr, debug.Stack())
				}
			}()

			g, err := gen.NewGenerator(spec, opt)
			require.NoError(t, err)
			require.NoError(t, g.WriteSource(genfs.CheckFS{}, "api"))
		})
		if len(opt.IgnoreNotImplemented) > 0 {
			// Check that all ignore rules are necessary.
			for _, feature := range ignore {
				if _, ok := notImplemented[feature]; !ok {
					t.Errorf("Ignore rule %q hasn't been used", feature)
				}
			}

			t.Run("Full", func(t *testing.T) {
				t.Skipf("Ignoring: %s", opt.IgnoreNotImplemented)
			})
		}
	}
}

func TestGenerate(t *testing.T) {
	t.Parallel()

	skipSets := map[string][]string{
		"autorest/additionalProperties.json": {
			"allOf",
		},
		"autorest/ApiManagementClient-openapi.json": {
			"allOf",
			"oauth2 security",
		},
		"autorest/lro.json": {
			"allOf",
		},
		"autorest/storage.json": {
			"allOf",
		},
		"autorest/xml-service.json": {
			"unsupported content types",
		},
		"autorest/xms-error-responses.json": {
			"allOf",
		},
		"2ch.yml": {
			"complex form schema",
		},
		"api.github.com.json": {
			"complex anyOf",
			"allOf",
			"discriminator inference",
			"sum types with same names",
			"sum type parameter",
			"unsupported content types",
		},
		"sample.json": {
			"enum format",
		},
		"manga.json":            {},
		"telegram_bot_api.json": {},
		"gotd_bot_api.json":     {},
		"k8s.json": {
			"unsupported content types",
		},
		"test_content_header_response.json": {
			"parameter content encoding",
		},
		"test_content_path_parameter.yml": {
			"parameter content encoding",
		},
		"petstore-expanded.yml": {
			"allOf",
		},
		"redoc/discriminator.json": {
			"unsupported content types",
		},
		"superset.json": {
			"allOf",
			"unsupported content types",
			"optional multipart file",
		},
	}

	testDataPath := "_testdata/positive"
	if err := fs.WalkDir(testdata, testDataPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return err
		}

		file := strings.TrimPrefix(path, testDataPath+"/")
		skip := skipSets[file]
		delete(skipSets, file)

		testName := file
		testName = strings.TrimSuffix(testName, ".json")
		testName = strings.TrimSuffix(testName, ".yml")
		testName = strings.TrimSuffix(testName, ".yaml")

		t.Run(testName, testGenerate(path, skip...))
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	// Check that skipSets needs update.
	if len(skipSets) > 0 {
		var schemas []string
		for k := range skipSets {
			schemas = append(schemas, k)
		}
		t.Fatalf("Schema ignore rules %+v have not been used.", schemas)
	}
}

func TestNegative(t *testing.T) {
	if err := fs.WalkDir(testdata, "_testdata/negative", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return err
		}
		_, file := filepath.Split(path)

		t.Run(strings.TrimSuffix(file, ".json"), func(t *testing.T) {
			a := require.New(t)
			data, err := testdata.ReadFile(path)
			a.NoError(err)

			spec, err := ogen.Parse(data)
			a.NoError(err)

			_, err = gen.NewGenerator(spec, gen.Options{})
			a.Error(err)
			t.Log(err.Error())
		})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
