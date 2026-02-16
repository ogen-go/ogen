package ogen_test

import (
	"net/url"
	"path"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen"
	"github.com/ogen-go/ogen/gen/genfs"
	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/location"
	"github.com/ogen-go/ogen/openapi/parser"
)

func testGenerate(t *testing.T, dir, filename string, data []byte, aliases ctAliases, ignore ...string) {
	t.Helper()
	t.Parallel()
	log := zaptest.NewLogger(t)

	spec, err := ogen.Parse(data)
	require.NoError(t, err)

	notImplemented := map[string]struct{}{}
	opt := gen.Options{
		Parser: gen.ParseOptions{
			InferSchemaType: true,
			File:            location.NewFile(filename, filename, data),
		},
		Generator: gen.GenerateOptions{
			IgnoreNotImplemented: ignore,
			NotImplementedHook: func(name string, err error) {
				notImplemented[name] = struct{}{}
			},
			ContentTypeAliases:         aliases,
			WildcardContentTypeDefault: ir.EncodingJSON,
		},
		Logger: log,
	}

	if filename == "file_reference.yml" { // HACK
		opt.Parser.AllowRemote = true
		opt.Parser.RootURL = &url.URL{
			Scheme: "file",
			Path:   "/" + path.Join(dir, filename),
		}
		opt.Parser.Remote = gen.RemoteOptions{
			ReadFile: func(p string) ([]byte, error) {
				p = strings.TrimPrefix(p, "/")
				return testdata.ReadFile(p)
			},
			URLToFilePath: func(u *url.URL) (string, error) {
				// By default, urlpath.URLToFilePath output depends on the OS.
				//
				// But we use virtual filesystem, so we should use the fs.FS path.
				if u.Path == "" {
					return u.Opaque, nil
				}
				return u.Path, nil
			},
		}
	}

	if path.Base(dir) == "convenient_errors" {
		require.NoError(t, opt.Generator.ConvenientErrors.Set("on"))
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

		if len(opt.Generator.IgnoreNotImplemented) > 0 {
			// Check that all ignore rules are necessary.
			for _, feature := range ignore {
				if _, ok := notImplemented[feature]; !ok {
					t.Errorf("Ignore rule %q hasn't been used", feature)
				}
			}
		}
	})
	t.Run("Full", func(t *testing.T) {
		t.Skipf("Ignoring: [%s]", strings.Join(opt.Generator.IgnoreNotImplemented, ", "))
	})
}

type ctAliases = map[string]ir.Encoding

func runPositive(root string,
	aliases map[string]ctAliases,
	skipSets map[string][]string,
) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()

		// Ensure that all skipSets schemas are present.
		for file := range skipSets {
			_, err := testdata.ReadFile(path.Join(root, file))
			require.NoErrorf(t, err, "skip file %s", file)
		}

		walkTestdata(t, root, func(t *testing.T, file string, data []byte) {
			dir := path.Dir(file)
			if parent := path.Base(dir); parent == "file_reference_external" {
				t.Skip("Special directory for testing remote references.")
				return
			}

			file = strings.TrimPrefix(file, root+"/")
			skip := skipSets[file]
			testGenerate(t, dir, file, data, aliases[file], skip...)
		})
	}
}

func TestGenerate(t *testing.T) {
	t.Run("Positive", runPositive("_testdata/positive", nil,
		map[string][]string{
			"sample.json": {
				"enum format",
			},
			"content_header_response.json": {
				"parameter content encoding",
			},
		}))

	t.Run("Examples", runPositive("_testdata/examples",
		map[string]ctAliases{
			"autorest/ApiManagementClient-openapi.json": {
				"text/json":                        ir.EncodingJSON,
				"application/vnd.swagger.doc+json": ir.EncodingJSON,
			},
			"api.github.com.json": {
				"text/x-markdown":            ir.EncodingTextPlain,
				"text/html":                  ir.EncodingTextPlain,
				"application/octocat-stream": ir.EncodingTextPlain,
				// FIXME(tdakkota): multiple response types makes wrapper cry about
				// 	type name conflict.
				// "application/vnd.github.v3.star+json": ir.EncodingJSON,
				"application/vnd.github.v3.object": ir.EncodingJSON,
				"application/scim+json":            ir.EncodingJSON,
			},
			"k8s.json": {
				"application/jwk-set+json":               ir.EncodingJSON,
				"application/merge-patch+json":           ir.EncodingJSON,
				"application/strategic-merge-patch+json": ir.EncodingJSON,
			},
			"problemjson.yml": {
				"application/problem+json": ir.EncodingProblemJSON,
			},
		},
		map[string][]string{
			"autorest/additionalProperties.json": {},
			"autorest/lro.json":                  {},
			"autorest/storage.json":              {},
			"autorest/xms-error-responses.json":  {},
			"2ch.yml":                            {},
			"api.github.com.json": {
				"complex anyOf",
				"discriminator inference",
				"sum types with same names",
				"array defaults",
				"type-based discrimination with same jxType",
			},
			"manga.json":               {},
			"telegram_bot_api.json":    {},
			"gotd_bot_api.json":        {},
			"k8s.json":                 {},
			"petstore-expanded.yml":    {},
			"problemjson.yml":          {},
			"redoc/discriminator.json": {},
			"swagger-petstore-1.0.27.yaml": {
				"nested objects in form parameters",
			},
		}))
}

// TestDuplicatePathsDifferentMethods tests that the generator correctly handles
// paths that normalize to the same structure but have different HTTP methods.
func TestDuplicatePathsDifferentMethods(t *testing.T) {
	log := zaptest.NewLogger(t)
	a := require.New(t)

	specYAML := `
openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /pets/{petId}:
    get:
      operationId: getPet
      parameters:
        - name: petId
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: OK
  /pets/{id}:
    post:
      operationId: createPet
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: OK
`
	spec, err := ogen.Parse([]byte(specYAML))
	a.NoError(err)

	opt := gen.Options{
		Parser: gen.ParseOptions{
			File: location.NewFile("test.yaml", "test.yaml", []byte(specYAML)),
		},
		Logger: log,
	}

	g, err := gen.NewGenerator(spec, opt)
	a.NoError(err)

	// Verify both operations were generated
	ops := g.Operations()
	a.Len(ops, 2)

	var foundGet, foundPost bool
	for _, op := range ops {
		switch op.Spec.OperationID {
		case "getPet":
			foundGet = true
			a.Equal("get", op.Spec.HTTPMethod)
		case "createPet":
			foundPost = true
			a.Equal("post", op.Spec.HTTPMethod)
		}
	}
	a.True(foundGet, "GET operation not found")
	a.True(foundPost, "POST operation not found")

	// Also verify that we can write the generated files without error
	err = g.WriteSource(genfs.CheckFS{}, "api")
	a.NoError(err)
}

func TestNegative(t *testing.T) {
	walkTestdata(t, "_testdata/negative", func(t *testing.T, file string, data []byte) {
		log := zaptest.NewLogger(t)

		a := require.New(t)
		dir, name := path.Split(file)

		spec, err := ogen.Parse(data)
		a.NoError(err)

		f := location.NewFile(name, name, data)
		_, err = parser.Parse(spec, parser.Settings{
			InferTypes: true,
			File:       f,
		})
		a.NoError(err, "If the error is related to parser, move this test to parser package testdata")

		opt := gen.Options{
			Parser: gen.ParseOptions{
				InferSchemaType: true,
				File:            f,
			},
			Logger: log,
		}
		t.Logf("Dir: %q, file: %q", dir, name)
		if strings.Contains(dir, "convenient_errors") {
			require.NoError(t, opt.Generator.ConvenientErrors.Set("on"))
		}

		_, err = gen.NewGenerator(spec, opt)
		a.Error(err)

		var buf strings.Builder
		if location.PrintPrettyError(&buf, true, err) {
			t.Log(buf.String())
		} else {
			t.Logf("%+v", err)
		}
	})
}
