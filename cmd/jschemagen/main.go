package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/ogen-go/ogen/gen"
	"github.com/ogen-go/ogen/gen/genfs"
)

func main() {
	var (
		specPath      = flag.String("schema", "", "Path to openapi spec file")
		targetFile    = flag.String("target", "output.go", "Path to target")
		packageName   = flag.String("package", os.Getenv("GOPACKAGE"), "Target package name")
		typeName      = flag.String("typename", "", "Root schema type name")
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

	dir, file := filepath.Split(filepath.Clean(*targetFile))
	if err := os.MkdirAll(dir, 0o750); err != nil {
		panic(err)
	}
	fs := genfs.FormattedSource{
		Root:   dir,
		Format: *performFormat,
	}
	if err := gen.GenerateSchema(data, fs, *typeName, file, *packageName); err != nil {
		panic(err)
	}
}
