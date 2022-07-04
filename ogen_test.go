package ogen_test

import (
	"embed"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen"
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

func TestParse(t *testing.T) {
	testcb := func(t *testing.T, file string, data []byte) {
		_, err := ogen.Parse(data)
		require.NoError(t, err)
	}

	t.Run("Positive", func(t *testing.T) {
		walkTestdata(t, "_testdata/positive", testcb)
	})
	t.Run("Examples", func(t *testing.T) {
		walkTestdata(t, "_testdata/examples", testcb)
	})
}
