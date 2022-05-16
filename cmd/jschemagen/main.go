package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/gen"
	"github.com/ogen-go/ogen/gen/genfs"
	"github.com/ogen-go/ogen/json"
	"github.com/ogen-go/ogen/jsonschema"
)

// StringArrayFlag is a string array flag.
type StringArrayFlag []string

// String implements fmt.Stringer.
func (i *StringArrayFlag) String() string {
	return strings.Join(*i, ",")
}

// Set implements flag.Value.
func (i *StringArrayFlag) Set(value string) error {
	*i = append(*i, strings.Split(value, ",")...)
	return nil
}

func run() error {
	var (
		specPath      = flag.String("schema", "", "Path to openapi spec file")
		targetFile    = flag.String("target", "output.go", "Path to target")
		packageName   = flag.String("package", os.Getenv("GOPACKAGE"), "Target package name")
		typeName      = flag.String("typename", "", "Root schema type name")
		performFormat = flag.Bool("format", true, "Perform code formatting")

		trimPrefixes StringArrayFlag
	)
	flag.Var(&trimPrefixes, "trim-prefixes", "Ref prefixes to trim")

	flag.Parse()
	if *specPath == "" {
		return errors.New("no spec provided")
	}

	data, err := os.ReadFile(*specPath)
	if err != nil {
		return errors.Wrap(err, "read file")
	}

	var rawSchema *jsonschema.RawSchema
	if err := json.Unmarshal(data, &rawSchema); err != nil {
		return errors.Wrap(err, "unmarshal")
	}
	p := jsonschema.NewParser(jsonschema.Settings{
		Resolver: jsonschema.NewRootResolver(data),
	})
	schema, err := p.Parse(rawSchema)
	if err != nil {
		return errors.Wrap(err, "parse")
	}

	dir, file := filepath.Split(filepath.Clean(*targetFile))
	if dir != "" {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return errors.Wrap(err, "create output directory")
		}
	}
	fs := genfs.FormattedSource{
		Root:   dir,
		Format: *performFormat,
	}

	if err := gen.GenerateSchema(schema, fs, gen.GenerateSchemaOptions{
		TypeName:   *typeName,
		FileName:   file,
		PkgName:    *packageName,
		TrimPrefix: trimPrefixes,
	}); err != nil {
		return errors.Wrap(err, "generate")
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
