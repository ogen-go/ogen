package main

import (
	"flag"
	"go/format"
	"os"
	"path/filepath"

	"github.com/ogen-go/ogen/gen"
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
	var (
		specPath    = flag.String("schema", "", "Path to openapi spec file")
		targetFile  = flag.String("target", "output.go", "Path to target")
		packageName = flag.String("package", os.Getenv("GOPACKAGE"), "Target package name")
		typeName = flag.String("typename", "", "Root schema type name")
		performFormat = flag.Bool("format", true, "perform code formatting")
	)

	flag.Parse()
	if *specPath == "" {
		panic("no spec provided")
	}
	data, err := os.ReadFile(*specPath)
	if err != nil {
		panic(err)
	}

	fs := formattedSource{
		Root:   "./",
		Format: *performFormat,
	}
	if err := gen.GenerateSchema(data, fs, *typeName, *targetFile, *packageName); err != nil {
		panic(err)
	}
}
