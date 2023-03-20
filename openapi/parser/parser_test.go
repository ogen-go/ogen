package parser_test

import (
	"embed"
	"fmt"
	"path"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	yaml "github.com/go-faster/yamlx"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/jsonpointer"
	"github.com/ogen-go/ogen/internal/location"
	"github.com/ogen-go/ogen/internal/testutil"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
	"github.com/ogen-go/ogen/openapi/parser"
)

//go:embed _testdata
var testdata embed.FS

func walkTestdata(t *testing.T, root string, cb func(t *testing.T, file string, data []byte)) {
	t.Helper()
	testutil.WalkTestdata(t, testdata, root, cb)
}

func TestNegative(t *testing.T) {
	walkTestdata(t, "_testdata/negative", func(t *testing.T, file string, data []byte) {
		a := require.New(t)
		_, name := path.Split(file)

		spec, err := ogen.Parse(data)
		a.NoError(err)

		_, err = parser.Parse(spec, parser.Settings{
			File: location.NewFile(name, file, data),
		})
		a.Error(err)

		var buf strings.Builder
		ok := location.PrintPrettyError(&buf, true, err)
		// Ensure that the error message is pretty printed.
		//
		// There should be a good reason to remove this line.
		a.True(ok)
		pretty := buf.String()
		a.NotEmpty(pretty)
		a.NotContains(pretty, location.BugLine)
		t.Log(pretty)
	})
}

func TestParserDeep(t *testing.T) {
	tests := []struct {
		name            string
		dir             string
		file            string
		expected        func() openapi.API
		expectedSpecErr string
		expectedAPIErr  string
	}{
		{
			name: "pet-tags",
			dir:  "_testdata/remotes/api",
			file: "pet-tags.yml",
			expected: func() openapi.API {
				petSchema := &jsonschema.Schema{
					Ref: jsonpointer.RefKey{
						Loc: "jsonschema://dummy",
						Ptr: "#/components/schemas/Pet",
					},
					Type: "object",
					Properties: []jsonschema.Property{
						{
							Name: "id",
							Schema: &jsonschema.Schema{
								Type:   "integer",
								Format: "int64",
								ExtraTags: map[string]string{
									"gorm":  "primaryKey",
									"valid": "customIdValidator",
								},
							},
							Required: true,
						},
						{
							Name: "name",
							Schema: &jsonschema.Schema{
								Type: "string",
								ExtraTags: map[string]string{
									"valid": "customNameValidator",
								},
							},
							Required: true,
						},
						{
							Name: "tag",
							Schema: &jsonschema.Schema{
								Type: "string",
							},
						},
					},
				}

				return openapi.API{
					Components: &openapi.Components{
						Schemas: map[string]*jsonschema.Schema{
							"Pet": petSchema,
						},
					},
				}
			},
		},
	}
	_ = tests

	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test %s", tt.name), func(t *testing.T) {
			a := require.New(t)

			filePath := path.Join(tt.dir, tt.file)
			fileData, fileErr := testdata.ReadFile(filePath)
			a.NoErrorf(fileErr, "load file %s", filePath)
			fileLocation := location.NewFile(tt.file, filePath, fileData)

			spec, specErr := ogen.Parse(fileData)
			a.NoErrorf(specErr, "parse file")

			var external jsonschema.ExternalResolver
			api, apiErr := parser.Parse(spec, parser.Settings{
				External:   external,
				File:       fileLocation,
				RootURL:    nil,
				InferTypes: false,
			})
			var apiErrText string
			if apiErr != nil {
				apiErrText = apiErr.Error()
			}
			a.Equal(tt.expectedAPIErr, apiErrText, "err")

			expected := tt.expected()
			for k, s := range expected.Components.Schemas {
				aSchema, has := api.Components.Schemas[k]
				if !has {
					continue
				}
				s.Locator = aSchema.Locator
				s.Source = aSchema.Source
				pMax := len(aSchema.Properties)
				if pMax > len(s.Properties) {
					pMax = len(s.Properties)
				}
				for p := 0; p < pMax; p++ {
					propActual := aSchema.Properties[p]
					s.Properties[p].Schema.Locator = propActual.Schema.Locator
					s.Properties[p].Schema.Source = propActual.Schema.Source
				}
			}

			checkDeep(t, "Components.Schemas", expected.Components.Schemas, api.Components.Schemas)
		})
	}
}

func checkDeep(t *testing.T, part string, expected, actual interface{}) {
	t.Helper()

	var err error
	var expectedY []byte
	if expected != nil {
		expectedY, err = yaml.Marshal(expected)
		require.NoError(t, err, part)
	}
	var actualY []byte
	if actual != nil {
		actualY, err = yaml.Marshal(actual)
		require.NoError(t, err, part)
	}

	// from github.com/stretchr/testify/assert/assertions.go:diff()
	var spewConfig = spew.ConfigState{
		Indent:                  " ",
		DisablePointerAddresses: true,
		DisableCapacities:       true,
		SortKeys:                true,
		DisableMethods:          true,
		MaxDepth:                10,
	}
	var spewConfigStringerEnabled = spew.ConfigState{
		Indent:                  " ",
		DisablePointerAddresses: true,
		DisableCapacities:       true,
		SortKeys:                true,
		MaxDepth:                10,
	}
	var e, a string

	switch reflect.TypeOf(expected) {
	case reflect.TypeOf(""):
		e = reflect.ValueOf(expected).String()
		a = reflect.ValueOf(actual).String()
	case reflect.TypeOf(time.Time{}):
		e = spewConfigStringerEnabled.Sdump(expected)
		a = spewConfigStringerEnabled.Sdump(actual)
	default:
		e = spewConfig.Sdump(expected)
		a = spewConfig.Sdump(actual)
	}

	require.Equal(t, expected, actual, fmt.Sprintf("%s\nEXPECTED:\n\n%s\n%s\nACTUAL:\n\n%s\n%s", part, expectedY, e, actualY, a))
}
