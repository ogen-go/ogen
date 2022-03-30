package gen

import (
	"embed"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/gen/genfs"
)

//go:embed _testdata/jsonschema
var testdata embed.FS

func TestGenerateSchema(t *testing.T) {
	require.NoError(t, fs.WalkDir(testdata, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return err
		}
		_, file := filepath.Split(path)

		input, err := fs.ReadFile(testdata, path)
		if err != nil {
			return err
		}
		t.Run(file, func(t *testing.T) {
			require.NoError(t, GenerateSchema(
				input,
				genfs.CheckFS{},
				"Type",
				"output.go",
				"output",
			))
		})

		return nil
	}))
}
