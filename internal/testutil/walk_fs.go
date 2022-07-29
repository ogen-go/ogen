// Package testutil contains helper functions for testing.
package testutil

import (
	"io/fs"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// WalkTestdata recursively walks through the root directory in given testdata FS
// and spawns a test for each file.
func WalkTestdata(t *testing.T, testdata fs.FS, root string, cb func(t *testing.T, file string, data []byte)) {
	t.Helper()

	dir, err := fs.ReadDir(testdata, root)
	require.NoError(t, err)

	for _, e := range dir {
		entryName := e.Name()
		filePath := path.Join(root, entryName)
		if e.IsDir() {
			t.Run(entryName, func(t *testing.T) {
				WalkTestdata(t, testdata, filePath, cb)
			})
			continue
		}

		testName := strings.TrimSuffix(entryName, ".json")
		testName = strings.TrimSuffix(testName, ".yml")
		testName = strings.TrimSuffix(testName, ".yaml")

		t.Run(testName, func(t *testing.T) {
			data, err := fs.ReadFile(testdata, filePath)
			require.NoError(t, err)
			cb(t, filePath, data)
		})
	}
}
