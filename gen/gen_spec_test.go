package gen

import (
	"encoding/json"
	"go/format"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-faster/yaml"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonschema"
)

// recordFS is a FileSystem that formats and records written files.
type recordFS map[string][]byte

func (fs recordFS) WriteFile(name string, content []byte) error {
	formatted, err := format.Source(content)
	if err != nil {
		return err
	}
	fs[name] = formatted
	return nil
}

func generateSpecTest(t *testing.T, spec *ogen.Spec, opts Options) (*Generator, recordFS) {
	t.Helper()

	g, err := NewGenerator(spec, opts)
	require.NoError(t, err)

	fs := recordFS{}
	require.NoError(t, g.WriteSource(fs, "api"))
	return g, fs
}

func enableFeatures(names ...string) *FeatureOptions {
	set := FeatureSet{}
	for _, name := range names {
		if err := set.Enable(name); err != nil {
			panic(err)
		}
	}
	return &FeatureOptions{Enable: set}
}

// petstoreSpec is a small document with integer response keys and a specification extension.
const petstoreSpec = `openapi: "3.0.0"
info:
  version: 1.0.0
  title: Swagger Petstore
  license:
    name: MIT
paths:
  /pets:
    get:
      x-ogen-operation-group: PetStore
      summary: List all pets
      operationId: listPets
      responses:
        200:
          description: A paged array of pets
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Pet"
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Error"
components:
  schemas:
    Pet:
      type: object
      required: [id, name]
      properties:
        id:
          type: integer
          format: int64
        name:
          type: string
    Error:
      type: object
      required: [code, message]
      properties:
        code:
          type: integer
          format: int32
        message:
          type: string
`

func parsePetstore(t *testing.T) *ogen.Spec {
	t.Helper()
	spec, err := ogen.Parse([]byte(petstoreSpec))
	require.NoError(t, err)
	return spec
}

func TestEmbeddedSpec(t *testing.T) {
	parse := parsePetstore

	t.Run("SwaggerUI", func(t *testing.T) {
		g, fs := generateSpecTest(t, parse(t), Options{
			Generator: GenerateOptions{
				Features: enableFeatures(ServerSwaggerUI.Name),
			},
		})
		require.NotNil(t, g.spec)
		require.False(t, g.spec.Expanded)
		require.Equal(t, "Swagger Petstore", g.spec.Title)

		// Embedded JSON is a valid document that keeps specification extensions.
		var decoded map[string]any
		require.NoError(t, json.Unmarshal([]byte(g.spec.JSON), &decoded))
		require.Contains(t, g.spec.JSON, `"x-ogen-operation-group": "PetStore"`)
		require.Contains(t, g.spec.JSON, `"200": {`)
		require.Contains(t, g.spec.YAML, `x-ogen-operation-group: PetStore`)

		// Embedded YAML parses to the same document.
		yamlSpec, err := ogen.Parse([]byte(g.spec.YAML))
		require.NoError(t, err)
		require.Equal(t, "Swagger Petstore", yamlSpec.Info.Title)

		src := string(fs["oas_spec_gen.go"])
		require.NotEmpty(t, src)
		require.Contains(t, src, "func SpecJSON() []byte")
		require.Contains(t, src, "func SpecYAML() []byte")
		require.Contains(t, src, "func SpecHandler(")
		require.Contains(t, src, "func SwaggerUIHandler(")
		require.Contains(t, src, "func WithDocsTitle(")
		require.Contains(t, src, `"github.com/ogen-go/ogen/swaggerui"`)

		cfg := string(fs["oas_cfg_gen.go"])
		require.Contains(t, cfg, "func WithDocs(")
		require.Contains(t, cfg, "newDocsHandler(true, opts...)")
		require.Contains(t, string(fs["oas_router_gen.go"]), "s.cfg.Docs")
	})

	t.Run("SpecOnly", func(t *testing.T) {
		_, fs := generateSpecTest(t, parse(t), Options{
			Generator: GenerateOptions{
				Features: enableFeatures(ServerSpec.Name),
			},
		})

		src := string(fs["oas_spec_gen.go"])
		require.Contains(t, src, "func SpecHandler(")
		require.NotContains(t, src, "SwaggerUIHandler")
		require.NotContains(t, src, "swaggerui")

		cfg := string(fs["oas_cfg_gen.go"])
		require.Contains(t, cfg, "func WithDocs(")
		require.Contains(t, cfg, "newDocsHandler(false, opts...)")
	})

	t.Run("Disabled", func(t *testing.T) {
		g, fs := generateSpecTest(t, parse(t), Options{})
		require.Nil(t, g.spec)
		require.NotContains(t, fs, "oas_spec_gen.go")
		require.NotContains(t, string(fs["oas_cfg_gen.go"]), "Docs")
		require.NotContains(t, string(fs["oas_router_gen.go"]), "Docs")
	})

	t.Run("NoServer", func(t *testing.T) {
		_, fs := generateSpecTest(t, parse(t), Options{
			Generator: GenerateOptions{
				Features: &FeatureOptions{
					Enable:  enableFeatures(PathsClient.Name, ServerSwaggerUI.Name).Enable,
					Disable: enableFeatures(PathsServer.Name).Enable,
				},
			},
		})

		// Standalone handlers are still generated.
		src := string(fs["oas_spec_gen.go"])
		require.Contains(t, src, "func SwaggerUIHandler(")
		require.NotContains(t, string(fs["oas_cfg_gen.go"]), "WithDocs")
	})

	t.Run("Expanded", func(t *testing.T) {
		_, fs := generateSpecTest(t, parse(t), Options{
			Generator: GenerateOptions{
				Features:   enableFeatures(ServerSpec.Name),
				SpecSource: SpecSourceExpanded,
			},
		})
		src := string(fs["oas_spec_gen.go"])
		require.Contains(t, src, "(expanded: external references are resolved)")
		require.Contains(t, src, `"title": "Swagger Petstore"`)
	})

	t.Run("Backquote", func(t *testing.T) {
		spec := parse(t)
		spec.Info.Description = "Uses `code` in description"
		spec.Raw = nil // Force marshaling of the modified spec.

		g, fs := generateSpecTest(t, spec, Options{
			Generator: GenerateOptions{
				Features: enableFeatures(ServerSpec.Name),
			},
		})
		require.Contains(t, g.spec.JSON, "Uses `code` in description")
		require.Contains(t, string(fs["oas_spec_gen.go"]), "Uses ` + \"`\" + `code` + \"`\" + ` in description")
	})

	t.Run("DSL", func(t *testing.T) {
		spec := ogen.NewSpec().
			SetOpenAPI("3.0.3").
			SetInfo(ogen.NewInfo().SetTitle("DSL").SetVersion("1.0.0")).
			AddPathItem("/ping", ogen.NewPathItem().
				SetGet(ogen.NewOperation().
					SetOperationID("ping").
					AddResponse("200", ogen.NewResponse().SetDescription("OK")),
				),
			)
		require.Nil(t, spec.Raw)

		g, fs := generateSpecTest(t, spec, Options{
			Generator: GenerateOptions{
				Features: enableFeatures(ServerSpec.Name),
			},
		})
		require.Equal(t, "DSL", g.spec.Title)
		require.Contains(t, g.spec.JSON, `"/ping"`)
		require.Contains(t, string(fs["oas_spec_gen.go"]), "func SpecHandler(")
	})
}

func TestEmbeddedSpecExternal(t *testing.T) {
	const root = `{
  "openapi": "3.0.3",
  "info": {"title": "external", "version": "1.0.0"},
  "paths": {
    "/pets": {
      "get": {
        "operationId": "listPets",
        "x-ogen-operation-group": "Pets",
        "responses": {
          "200": {
            "description": "OK",
            "content": {"application/json": {"schema": {"$ref": "schemas.json#/Pet"}}}
          }
        }
      }
    }
  }
}`
	const external = `{
  "Pet": {"type": "object", "properties": {"name": {"type": "string"}}}
}`

	parse := func(t *testing.T) (*ogen.Spec, Options) {
		files := map[string]string{
			"root.json":    root,
			"schemas.json": external,
		}
		opts := Options{
			Parser: ParseOptions{
				AllowRemote: true,
			},
		}
		data, err := opts.SetLocation("root.json", jsonschema.ExternalOptions{
			ReadFile: func(p string) ([]byte, error) {
				content, ok := files[filepath.Base(p)]
				require.True(t, ok, p)
				return []byte(content), nil
			},
		})
		require.NoError(t, err)

		spec, err := ogen.Parse(data)
		require.NoError(t, err)
		return spec, opts
	}

	t.Run("Auto", func(t *testing.T) {
		spec, opts := parse(t)
		opts.Generator.Features = enableFeatures(ServerSpec.Name)
		g, _ := generateSpecTest(t, spec, opts)

		require.True(t, g.spec.Expanded)
		require.NotContains(t, g.spec.JSON, "schemas.json#")
		require.Contains(t, g.spec.JSON, `"name"`)
	})

	t.Run("Raw", func(t *testing.T) {
		spec, opts := parse(t)
		opts.Generator.Features = enableFeatures(ServerSpec.Name)
		opts.Generator.SpecSource = SpecSourceRaw
		g, _ := generateSpecTest(t, spec, opts)

		require.False(t, g.spec.Expanded)
		require.Contains(t, g.spec.JSON, "schemas.json#/Pet")
		require.Contains(t, g.spec.JSON, "x-ogen-operation-group")
	})

	t.Run("NoExternal", func(t *testing.T) {
		g, _ := generateSpecTest(t, parsePetstore(t), Options{
			Parser: ParseOptions{AllowRemote: true},
			Generator: GenerateOptions{
				Features: enableFeatures(ServerSpec.Name),
			},
		})
		require.False(t, g.spec.Expanded, "allow_remote alone must not force expansion")
	})
}

func TestStringKeys(t *testing.T) {
	var n yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte("responses:\n  200: &ok {description: OK}\n  201: *ok\n  true: yes\n"), &n))

	c := stringKeys(&n)

	// Original node is not modified.
	require.Equal(t, "!!int", n.Content[0].Content[1].Content[0].Tag)
	responses := c.Content[0].Content[1]
	for i := 0; i < len(responses.Content); i += 2 {
		require.Equal(t, "!!str", responses.Content[i].Tag)
	}
	// Values are not modified.
	require.Equal(t, yaml.MappingNode, responses.Content[1].Kind)
	require.Equal(t, yaml.AliasNode, responses.Content[3].Kind)
	require.Equal(t, "!!str", responses.Content[5].Tag)
	require.Equal(t, "yes", responses.Content[5].Value)

	var buf strings.Builder
	data, err := yaml.Marshal(c)
	require.NoError(t, err)
	buf.Write(data)
	require.Contains(t, buf.String(), `"200":`)
}

func TestSpecSourceYAML(t *testing.T) {
	var o GenerateOptions
	require.NoError(t, yaml.Unmarshal([]byte("spec_source: expanded"), &o))
	require.Equal(t, SpecSourceExpanded, o.SpecSource)

	require.Error(t, yaml.Unmarshal([]byte("spec_source: foo"), &o))
}
