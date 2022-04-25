package ogen_test

import (
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen"
)

func TestNegative(t *testing.T) {
	if err := fs.WalkDir(testdata, "_testdata/negative", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return err
		}
		_, file := filepath.Split(path)

		t.Run(strings.TrimSuffix(file, ".json"), func(t *testing.T) {
			a := require.New(t)
			data, err := testdata.ReadFile(path)
			a.NoError(err)

			spec, err := ogen.Parse(data)
			a.NoError(err)

			_, err = gen.NewGenerator(spec, gen.Options{})
			a.Error(err)
			t.Log(err.Error())
		})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
