package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDetectGoVersion(t *testing.T) {
	a := require.New(t)
	root := t.TempDir()

	// Target dir may not exist yet, the closest go.mod is used.
	target := filepath.Join(root, "internal", "api")

	a.NoError(os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/m\n\ngo 1.27.1\n"), 0o600))
	v, err := detectGoVersion(target)
	a.NoError(err)
	a.Equal("1.27.1", v)

	a.NoError(os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/m\n"), 0o600))
	v, err = detectGoVersion(target)
	a.NoError(err)
	a.Empty(v)
}
