// Binary ogen generates go source code from OAS.
package main

import (
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/gen"
)

type formattedSource struct {
	Format bool
	Root   string
}

func (t formattedSource) WriteFile(name string, content []byte) error {
	out := content
	if t.Format {
		buf, err := format.Source(content)
		if err != nil {
			return err
		}
		out = buf
	}
	return os.WriteFile(filepath.Join(t.Root, name), out, 0600)
}

func main() {
	specPath := flag.String("schema", "", "Path to openapi spec file")
	targetDir := flag.String("target", "api", "Path to target dir")
	packageName := flag.String("package", "api", "Target package name")
	performFormat := flag.Bool("format", true, "perform code formatting")
	clean := flag.Bool("clean", false, "Clean generated files before generation")
	flag.Parse()
	if *specPath == "" {
		panic("no spec provided")
	}
	f, err := os.Open(*specPath)
	if err != nil {
		panic(err)
	}
	defer func() { _ = f.Close() }()

	spec, err := ogen.Parse(f)
	if err != nil {
		panic(err)
	}
	files, err := os.ReadDir(*targetDir)
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	if os.IsNotExist(err) {
		if err := os.Mkdir(*targetDir, 0750); err != nil {
			panic(err)
		}
	}
	if *clean {
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			name := f.Name()
			if !strings.HasSuffix(name, "_gen.go") {
				continue
			}
			if !strings.HasPrefix(name, "openapi") {
				continue
			}
			if err := os.Remove(filepath.Join(*targetDir, name)); err != nil {
				panic(err)
			}
		}
	}

	fs := formattedSource{
		Root:   *targetDir,
		Format: *performFormat,
	}
	g, err := gen.NewGenerator(spec)
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}
	if err := g.WriteSource(fs, *packageName); err != nil {
		panic(fmt.Sprintf("%+v", err))
	}
}
