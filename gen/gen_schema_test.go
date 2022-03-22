package gen

import (
	"embed"
	"go/format"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TODO: Create validationFs.
type fmtFs struct{}

func (n fmtFs) WriteFile(baseName string, source []byte) error {
	_, err := format.Source(source)
	return err
}

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
				fmtFs{},
				"Type",
				"output.go",
				"output",
			))
		})

		return nil
	}))
}
