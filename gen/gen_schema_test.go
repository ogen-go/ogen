package gen

import (
	"embed"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-faster/errors"
	yaml "github.com/go-faster/yamlx"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/ogen-go/ogen/gen/genfs"
	"github.com/ogen-go/ogen/jsonschema"
)

//go:embed _testdata/jsonschema
var testdata embed.FS

func TestGenerateSchema(t *testing.T) {
	logger := zaptest.NewLogger(t)

	require.NoError(t, fs.WalkDir(testdata, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return err
		}
		_, file := filepath.Split(path)

		data, err := fs.ReadFile(testdata, path)
		if err != nil {
			return err
		}

		var root yaml.Node
		if err := yaml.Unmarshal(data, &root); err != nil {
			return errors.Wrap(err, "parse yaml")
		}
		p := jsonschema.NewParser(jsonschema.Settings{
			Resolver: jsonschema.NewRootResolver(&root),
		})

		var rawSchema jsonschema.RawSchema
		if err := root.Decode(&rawSchema); err != nil {
			return errors.Wrap(err, "unmarshal")
		}
		schema, err := p.Parse(&rawSchema)
		if err != nil {
			return errors.Wrap(err, "parse")
		}

		t.Run(strings.TrimSuffix(file, ".json"), func(t *testing.T) {
			require.NoError(t, GenerateSchema(
				schema,
				genfs.CheckFS{},
				GenerateSchemaOptions{
					Logger: logger,
				},
			))
		})

		return nil
	}))
}
