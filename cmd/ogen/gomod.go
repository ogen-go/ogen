package main

import (
	"os"
	"path/filepath"

	"github.com/go-faster/errors"
	"golang.org/x/mod/modfile"
)

// detectGoVersion returns the Go version of the module containing dir, as set
// by the go directive of its go.mod.
//
// The dir may not exist yet. Empty string is returned if there is no go.mod or
// it has no go directive.
func detectGoVersion(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		p := filepath.Join(dir, "go.mod")
		data, err := os.ReadFile(p) //#nosec G304,G703 -- go.mod lookup in the user-provided target dir
		switch {
		case err == nil:
			f, err := modfile.ParseLax(p, data, nil)
			if err != nil {
				return "", errors.Wrapf(err, "parse %s", p)
			}
			if f.Go == nil {
				return "", nil
			}
			return f.Go.Version, nil
		case !errors.Is(err, os.ErrNotExist):
			return "", errors.Wrapf(err, "read %s", p)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", nil
		}
		dir = parent
	}
}
