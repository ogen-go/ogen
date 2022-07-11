package ogen_test

import (
	"path"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen"
	"github.com/ogen-go/ogen/gen/genfs"
)

func testGenerate(t *testing.T, data []byte, ignore ...string) {
	t.Helper()
	t.Parallel()
	log := zaptest.NewLogger(t)

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
			t.Skipf("Ignoring: [%s]", strings.Join(opt.IgnoreNotImplemented, ", "))
		})
	}
}

func runPositive(root string, skipSets map[string][]string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()

		// Ensure that all skipSets schemas are present.
		for file := range skipSets {
			_, err := testdata.ReadFile(path.Join(root, file))
			require.NoErrorf(t, err, "skip file %s", file)
		}

		walkTestdata(t, root, func(t *testing.T, file string, data []byte) {
			file = strings.TrimPrefix(file, root+"/")
			skip := skipSets[file]
			testGenerate(t, data, skip...)
		})
	}
}

func TestGenerate(t *testing.T) {
	t.Run("Positive", runPositive("_testdata/positive", map[string][]string{
		"sample.json": {
			"enum format",
		},
		"content_header_response.json": {
			"parameter content encoding",
		},
		"content_path_parameter.yml": {
			"parameter content encoding",
		},
	}))

	t.Run("Examples", runPositive("_testdata/examples", map[string][]string{
		"autorest/additionalProperties.json": {},
		"autorest/ApiManagementClient-openapi.json": {
			"oauth2 security",
		},
		"autorest/lro.json":                 {},
		"autorest/storage.json":             {},
		"autorest/xms-error-responses.json": {},
		"2ch.yml":                           {},
		"api.github.com.json": {
			"complex anyOf",
			"discriminator inference",
			"sum types with same names",
			"sum type parameter",
			"unsupported content types",
		},
		"manga.json":            {},
		"telegram_bot_api.json": {},
		"gotd_bot_api.json":     {},
		"k8s.json": {
			"unsupported content types",
		},
		"petstore-expanded.yml": {},
		"redoc/discriminator.json": {
			"unsupported content types",
		},
	}))
}

func TestNegative(t *testing.T) {
	walkTestdata(t, "_testdata/negative", func(t *testing.T, file string, data []byte) {
		a := require.New(t)
		_, name := path.Split(file)

		spec, err := ogen.Parse(data)
		a.NoError(err)

		_, err = gen.NewGenerator(spec, gen.Options{
			Filename: name,
		})
		a.Error(err)
		t.Logf("%+v", err)
	})
}
