package parser_test

import (
	"embed"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/location"
	"github.com/ogen-go/ogen/openapi/parser"
)

//go:embed _testdata
var testdata embed.FS

func walkTestdata(t *testing.T, root string, cb func(t *testing.T, file string, data []byte)) {
	t.Helper()

	dir, err := testdata.ReadDir(root)
	require.NoError(t, err)

	for _, e := range dir {
		entryName := e.Name()
		filePath := path.Join(root, entryName)
		if e.IsDir() {
			t.Run(entryName, func(t *testing.T) {
				walkTestdata(t, filePath, cb)
			})
			continue
		}

		testName := strings.TrimSuffix(entryName, ".json")
		testName = strings.TrimSuffix(testName, ".yml")
		testName = strings.TrimSuffix(testName, ".yaml")

		t.Run(testName, func(t *testing.T) {
			data, err := testdata.ReadFile(filePath)
			require.NoError(t, err)
			cb(t, filePath, data)
		})
	}
}

func TestNegative(t *testing.T) {
	walkTestdata(t, "_testdata/negative", func(t *testing.T, file string, data []byte) {
		a := require.New(t)
		_, name := path.Split(file)

		spec, err := ogen.Parse(data)
		a.NoError(err)

		_, err = parser.Parse(spec, parser.Settings{
			Filename: name,
		})
		a.Error(err)

		var buf strings.Builder
		ok := location.PrintPrettyError(&buf, name, data, err)
		// Ensure that the error message is pretty printed.
		//
		// There should be a good reason to remove this line.
		a.True(ok)
		t.Log(buf.String())
	})
}
