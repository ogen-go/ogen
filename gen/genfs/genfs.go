// Package genfs contains gen.FileSystem implementations.
package genfs

import (
	"go/format"
	"os"
	"path/filepath"
)

// FormattedSource is gen.FileSystem implementation that format and writes Go sources.
type FormattedSource struct {
	Format bool
	Root   string
}

// WriteFile implements gen.FileSystem.
func (t FormattedSource) WriteFile(name string, content []byte) error {
	out := content
	if t.Format {
		buf, err := format.Source(content)
		if err != nil {
			return err
		}
		out = buf
	}
	return os.WriteFile(filepath.Join(t.Root, name), out, 0o644)
}
