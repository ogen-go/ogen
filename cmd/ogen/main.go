// Binary ogen generates go source code from OAS.
package main

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-faster/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen"
	"github.com/ogen-go/ogen/gen/genfs"
	"github.com/ogen-go/ogen/internal/ogenzap"
)

func cleanDir(targetDir string, files []os.DirEntry) error {
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		if !strings.HasSuffix(name, "_gen.go") || !strings.HasSuffix(name, "_gen_test.go") {
			continue
		}
		if !(strings.HasPrefix(name, "openapi") || strings.HasPrefix(name, "oas")) {
			continue
		}
		if err := os.Remove(filepath.Join(targetDir, name)); err != nil {
			return err
		}
	}
	return nil
}

func generate(specPath, packageName string, fs gen.FileSystem, opts gen.Options) error {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return err
	}

	spec, err := ogen.Parse(data)
	if err != nil {
		return errors.Wrap(err, "parse spec")
	}

	g, err := gen.NewGenerator(spec, opts)
	if err != nil {
		return errors.Wrap(err, "build IR")
	}

	if err := g.WriteSource(fs, packageName); err != nil {
		return errors.Wrap(err, "write")
	}

	return nil
}

func run() error {
	var (
		targetDir         = flag.String("target", "api", "Path to target dir")
		packageName       = flag.String("package", "api", "Target package name")
		inferTypes        = flag.Bool("infer-types", false, "Infer schema types, if type is not defined explicitly")
		performFormat     = flag.Bool("format", true, "Perform code formatting")
		verbose           = flag.Bool("v", false, "Enable verbose logging")
		logLevel          = zap.LevelFlag("loglevel", zapcore.InfoLevel, "Zap logging level")
		clean             = flag.Bool("clean", false, "Clean generated files before generation")
		generateTests     = flag.Bool("generate-tests", false, "Generate tests encode-decode/based on schema examples")
		skipTestsRegex    = flag.String("skip-tests", "", "Skip tests matched by regex")
		skipUnimplemented = flag.Bool("skip-unimplemented", false, "Disables generation of UnimplementedHandler")

		debugIgnoreNotImplemented = flag.String("debug.ignoreNotImplemented", "",
			"Ignore methods having functionality which is not implemented "+
				"(all, oneOf, anyOf, allOf, nullable types, complex parameter types)")
		debugNoerr = flag.Bool("debug.noerr", false, "Ignore all errors")
	)

	var (
		filterPath    *regexp.Regexp
		filterMethods []string
	)
	flag.Func("filter-path", "Filter operations by path regex", func(s string) (err error) {
		filterPath, err = regexp.Compile(s)
		return err
	})
	flag.Func("filter-methods", "Filter operations by HTTP methods (comma-separated)", func(s string) error {
		for _, m := range strings.Split(s, ",") {
			m = strings.TrimSpace(m)
			if m == "" {
				continue
			}
			filterMethods = append(filterMethods, m)
		}
		return nil
	})

	flag.Parse()

	specPath := flag.Arg(0)
	if flag.NArg() == 0 || specPath == "" {
		return errors.New("no spec provided")
	}

	switch files, err := os.ReadDir(*targetDir); {
	case os.IsNotExist(err):
		if err := os.MkdirAll(*targetDir, 0750); err != nil {
			return err
		}
	case err != nil:
		return err
	default:
		if *clean {
			if err := cleanDir(*targetDir, files); err != nil {
				return errors.Wrap(err, "clean")
			}
		}
	}

	logger, err := ogenzap.Create(*logLevel, *verbose)
	if err != nil {
		return err
	}
	defer func() {
		_ = logger.Sync()
	}()

	opts := gen.Options{
		VerboseRoute:         *verbose,
		GenerateExampleTests: *generateTests,
		SkipTestRegex:        nil, // Set below.
		SkipUnimplemented:    *skipUnimplemented,
		InferSchemaType:      *inferTypes,
		Filters: gen.Filters{
			PathRegex: filterPath,
			Methods:   filterMethods,
		},
		IgnoreNotImplemented: strings.Split(*debugIgnoreNotImplemented, ","),
		NotImplementedHook:   nil,
		Logger:               logger,
	}
	if expr := *skipTestsRegex; expr != "" {
		r, err := regexp.Compile(expr)
		if err != nil {
			return errors.Wrap(err, "skipTestsRegex")
		}
		opts.SkipTestRegex = r
	}
	if *debugNoerr {
		opts.IgnoreNotImplemented = []string{"all"}
	}

	fs := genfs.FormattedSource{
		Root:   *targetDir,
		Format: *performFormat,
	}
	if err := generate(specPath, *packageName, fs, opts); err != nil {
		return errors.Wrap(err, "generate")
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
