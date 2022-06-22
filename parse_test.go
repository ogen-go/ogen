package ogen_test

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
)

func TestParse(t *testing.T) {
	testDataPath := "_testdata/positive"
	if err := fs.WalkDir(testdata, testDataPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return err
		}
		file := strings.TrimPrefix(path, testDataPath+"/")

		testName := file
		testName = strings.TrimSuffix(testName, ".json")
		testName = strings.TrimSuffix(testName, ".yml")
		testName = strings.TrimSuffix(testName, ".yaml")

		t.Run(testName, func(t *testing.T) {
			a := require.New(t)

			data, err := testdata.ReadFile(path)
			a.NoError(err)

			_, err = ogen.Parse(data)
			a.NoError(err)
		})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
