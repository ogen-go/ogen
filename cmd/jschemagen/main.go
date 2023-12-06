package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-faster/errors"
	"github.com/go-faster/yaml"

	"github.com/ogen-go/ogen/gen"
	"github.com/ogen-go/ogen/gen/genfs"
	"github.com/ogen-go/ogen/internal/ogenzap"
	"github.com/ogen-go/ogen/internal/urlpath"
	"github.com/ogen-go/ogen/jsonpointer"
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

func inferFileName(
	targetFile *string,
	typeName string,
	rawSchema jsonschema.RawSchema,
	trimPrefixes StringArrayFlag,
) {
	// Output file already set.
	if *targetFile != "" {
		return
	}

	if typeName == "" {
		if ref := rawSchema.Ref; ref != "" {
			for _, prefix := range trimPrefixes {
				ref = strings.TrimPrefix(ref, prefix)
			}
			typeName = ref
		}
	}

	typeName = strings.ToLower(typeName)
	// Check that type name contains only valid path characters.
	if regexp.MustCompile(`^\w+$`).MatchString(typeName) {
		*targetFile = typeName + "_json.go"
	}

	if *targetFile == "" {
		*targetFile = "output.go"
	}
}

func run() error {
	var (
		targetFile    = flag.String("target", "", "Path to target")
		packageName   = flag.String("package", os.Getenv("GOPACKAGE"), "Target package name")
		typeName      = flag.String("typename", "", "Root schema type name")
		inferTypes    = flag.Bool("infer-types", false, "Infer schema types, if type is not defined explicitly")
		performFormat = flag.Bool("format", true, "Perform code formatting")
		trimPrefixes  = StringArrayFlag{"#/definitions/", "#/$defs/"}
		logOptions    ogenzap.Options
	)
	flag.Var(&trimPrefixes, "trim-prefixes", "Ref prefixes to trim")
	logOptions.RegisterFlags(flag.CommandLine)

	flag.Parse()
	specPath := flag.Arg(0)
	if flag.NArg() < 1 || specPath == "" {
		return errors.New("no spec provided")
	}
	specPath = filepath.Clean(specPath)
	abs, err := filepath.Abs(specPath)
	if err != nil {
		return err
	}
	rootURL, err := urlpath.URLFromFilePath(abs)
	if err != nil {
		return errors.Wrap(err, "convert file path to url")
	}

	data, err := os.ReadFile(specPath)
	if err != nil {
		return errors.Wrap(err, "read file")
	}

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return errors.Wrap(err, "parse yaml")
	}
	p := jsonschema.NewParser(jsonschema.Settings{
		Resolver:   jsonschema.NewRootResolver(&root),
		InferTypes: *inferTypes,
	})

	var rawSchema jsonschema.RawSchema
	if err := root.Decode(&rawSchema); err != nil {
		return errors.Wrap(err, "unmarshal")
	}
	schema, err := p.Parse(&rawSchema, jsonpointer.NewResolveCtx(rootURL, jsonpointer.DefaultDepthLimit))
	if err != nil {
		return errors.Wrap(err, "parse")
	}

	inferFileName(targetFile, *typeName, rawSchema, trimPrefixes)
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

	if *packageName == "" {
		*packageName = "output"
	}

	logger, err := ogenzap.Create(logOptions)
	if err != nil {
		return err
	}
	defer func() {
		_ = logger.Sync()
	}()
	if err := gen.GenerateSchema(schema, fs, gen.GenerateSchemaOptions{
		TypeName:   *typeName,
		FileName:   file,
		PkgName:    *packageName,
		TrimPrefix: trimPrefixes,
		Logger:     logger,
	}); err != nil {
		return errors.Wrap(err, "generate")
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
