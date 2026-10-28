package gen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/jsonschema"
)

type captureFS struct {
	files map[string][]byte
}

func (c *captureFS) WriteFile(baseName string, source []byte) error {
	if c.files == nil {
		c.files = map[string][]byte{}
	}
	c.files[baseName] = source
	return nil
}

func TestGenerateSchemaConstURICompiles(t *testing.T) {
	a := require.New(t)

	const uri = "https://example.com/context.jsonld"
	schema := &jsonschema.Schema{
		Type: jsonschema.Object,
		Properties: []jsonschema.Property{
			{
				Name: "@context",
				Schema: &jsonschema.Schema{
					Type:     jsonschema.String,
					Format:   "uri",
					Const:    uri,
					ConstSet: true,
				},
				Required: true,
			},
			{
				Name: "homepage",
				Schema: &jsonschema.Schema{
					Type:     jsonschema.String,
					Format:   "uri",
					Const:    "https://example.com/",
					ConstSet: true,
				},
			},
		},
	}

	fs := &captureFS{}
	err := GenerateSchema(schema, fs, GenerateSchemaOptions{
		TypeName: "Source",
		FileName: "output.gen.go",
		PkgName:  "output",
	})
	a.NoError(err)

	src, ok := fs.files["output.gen.go"]
	a.True(ok, "expected generated output.gen.go")
	got := string(src)

	// Regression for ogen-go/ogen#1751: const uri was passed as a string to
	// json.EncodeURI, which expects url.URL, so generated code did not compile.
	a.NotContains(got, `json.EncodeURI(e, "`+uri+`")`)
	a.NotContains(got, `json.EncodeURI(e, "https://example.com/")`)
	a.Contains(got, `e.Str("`+uri+`")`)
	a.Contains(got, `e.Str("https://example.com/")`)
}
