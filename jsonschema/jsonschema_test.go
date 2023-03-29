package jsonschema_test

import (
	"embed"
	"net/url"
	"path"
	"strings"
	"testing"

	"github.com/go-faster/yaml"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/internal/jsonpointer"
	"github.com/ogen-go/ogen/internal/location"
	"github.com/ogen-go/ogen/internal/testutil"
	"github.com/ogen-go/ogen/jsonschema"
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
		t.Log(pretty)
	})
}
