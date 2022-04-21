package gen

import (
	"embed"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/gen/genfs"
	"github.com/ogen-go/ogen/json"
	"github.com/ogen-go/ogen/jsonschema"
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

		var rawSchema *jsonschema.RawSchema
		if err := json.Unmarshal(input, &rawSchema); err != nil {
			return errors.Wrap(err, "unmarshal")
		}
		p := jsonschema.NewParser(jsonschema.Settings{
			Resolver: jsonschema.NewRootResolver(input),
		})
		schema, err := p.Parse(rawSchema)
		if err != nil {
			return errors.Wrap(err, "parse")
		}

		t.Run(strings.TrimSuffix(file, ".json"), func(t *testing.T) {
			require.NoError(t, GenerateSchema(
				schema,
				genfs.CheckFS{},
				GenerateSchemaOptions{},
			))
		})

		return nil
	}))
}
