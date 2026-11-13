package jsonschema_test

import (
	"embed"
	"net/url"
	"path"
	"strings"
	"testing"

	"github.com/go-faster/yaml"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/internal/testutil"
	"github.com/ogen-go/ogen/jsonpointer"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/location"
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

		var schema jsonschema.RawSchema
		err := yaml.Unmarshal(data, &schema)
		a.NoError(err)

		p := jsonschema.NewParser(jsonschema.Settings{
			File: location.NewFile(name, file, data),
		})
		_, err = p.Parse(&schema, jsonpointer.NewResolveCtx(&url.URL{Path: "/" + file}, jsonpointer.DefaultDepthLimit))
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
		t.Logf("\n%s", pretty)
	})
}

func TestMinMaxConflictError(t *testing.T) {
	tests := []struct {
		file string
		err  string
	}{
		{
			"_testdata/negative/array/invalid_minItems.json",
			"at invalid_minItems.json:1:1: " +
				"- at invalid_minItems.json:4:15: minItems (10) is greater than maxItems (1)\n" +
				"- at invalid_minItems.json:5:15: \n",
		},
		{
			"_testdata/negative/object/invalid_minProperties.json",
			"at invalid_minProperties.json:1:1: " +
				"- at invalid_minProperties.json:4:20: minProperties (10) is greater than maxProperties (1)\n" +
				"- at invalid_minProperties.json:5:20: \n",
		},
		{
			"_testdata/negative/string/invalid_minLength.json",
			"at invalid_minLength.json:1:1: " +
				"- at invalid_minLength.json:4:16: minLength (10) is greater than maxLength (1)\n" +
				"- at invalid_minLength.json:5:16: \n",
		},
	}

	for _, tt := range tests {
		_, name := path.Split(tt.file)

		t.Run(strings.TrimSuffix(name, ".json"), func(t *testing.T) {
			a := require.New(t)

			data, err := testdata.ReadFile(tt.file)
			a.NoError(err)

			var schema jsonschema.RawSchema
			a.NoError(yaml.Unmarshal(data, &schema))

			p := jsonschema.NewParser(jsonschema.Settings{
				File: location.NewFile(name, tt.file, data),
			})
			_, err = p.Parse(&schema, jsonpointer.NewResolveCtx(&url.URL{Path: "/" + tt.file}, jsonpointer.DefaultDepthLimit))
			a.EqualError(err, tt.err)
		})
	}
}
