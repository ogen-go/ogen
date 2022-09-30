// Binary ogen generates go source code from OAS.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen"
	"github.com/ogen-go/ogen/gen/genfs"
	"github.com/ogen-go/ogen/internal/location"
	"github.com/ogen-go/ogen/internal/ogenversion"
	"github.com/ogen-go/ogen/internal/ogenzap"
)

func cleanDir(targetDir string, files []os.DirEntry) error {
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		if !(strings.HasSuffix(name, "_gen.go") || strings.HasSuffix(name, "_gen_test.go")) {
			continue
		}
		if !(strings.HasPrefix(name, "openapi") || strings.HasPrefix(name, "oas")) {
			continue
		}
		// Do not return error if file does not exist.
		if err := os.Remove(filepath.Join(targetDir, name)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func generate(data []byte, packageName, targetDir string, clean bool, opts gen.Options) error {
	log := opts.Logger

	spec, err := ogen.Parse(data)
	if err != nil {
		return errors.Wrap(err, "parse spec")
	}

	start := time.Now()
	g, err := gen.NewGenerator(spec, opts)
	if err != nil {
		return errors.Wrap(err, "build IR")
	}
	log.Debug("Build IR", zap.Duration("took", time.Since(start)))

	// Clean target dir only after flag parsing, spec parsing and IR building.
	switch files, err := os.ReadDir(targetDir); {
	case os.IsNotExist(err):
		if err := os.MkdirAll(targetDir, 0o750); err != nil {
			return err
		}
	case err != nil:
		return err
	default:
		if clean {
			if err := cleanDir(targetDir, files); err != nil {
				return errors.Wrap(err, "clean")
			}
		}
	}

	fs := genfs.FormattedSource{
		// FIXME(tdakkota): write source uses imports.Process which also uses go/format.
		// 	So, there is no reason to format source twice or provide a flag to disable formatting.
		Format: false,
		Root:   targetDir,
	}
	start = time.Now()
	if err := g.WriteSource(fs, packageName); err != nil {
		return errors.Wrap(err, "write")
	}
	log.Debug("Write", zap.Duration("took", time.Since(start)))

	return nil
}

func handleGenerateError(w io.Writer, specPath string, data []byte, err error) (r bool) {
	defer func() {
		// Add trailing newline to the error message if error is handled.
		if r {
			_, _ = fmt.Fprintln(w)
		}
	}()

	if location.PrintPrettyError(w, specPath, data, err) {
		return true
	}

	if notImplErr, ok := errors.Into[*gen.ErrNotImplemented](err); ok {
		_, _ = fmt.Fprintf(w, `
Feature %[1]q is not implemented yet.
Try to run ogen with --debug.ignoreNotImplemented %[1]q or with --debug.noerr to skip unsupported operations.
`, notImplErr.Name)
		return true
	}

	if ctErr, ok := errors.Into[*gen.ErrUnsupportedContentTypes](err); ok {
		_, _ = fmt.Fprintf(w, `
Content types [%s] are unsupported.
Try to run ogen with --debug.ignoreNotImplemented %q or with --debug.noerr to skip unsupported operations.
Also, you can use --ct-alias to map content types to supported ones.
`,
			strings.Join(ctErr.ContentTypes, ", "),
			"unsupported content type",
		)
		return true
	}

	return false
}

func run() error {
	set := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	set.Usage = func() {
		_, toolName := filepath.Split(os.Args[0])
		_, _ = fmt.Fprintf(set.Output(), "Usage: %s [options] <spec>\n", toolName)
		set.PrintDefaults()
	}

	var (
		// Generator options.
		targetDir         = set.String("target", "api", "Path to target dir")
		packageName       = set.String("package", "api", "Target package name")
		inferTypes        = set.Bool("infer-types", false, "Infer schema types, if type is not defined explicitly")
		clean             = set.Bool("clean", false, "Clean generated files before generation")
		generateTests     = set.Bool("generate-tests", false, "Generate tests encode-decode/based on schema examples")
		allowRemote       = set.Bool("allow-remote", false, "Enables remote references resolving")
		skipTestsRegex    = set.String("skip-tests", "", "Skip tests matched by regex")
		skipUnimplemented = set.Bool("skip-unimplemented", false, "Disables generation of UnimplementedHandler")
		noClient          = set.Bool("no-client", false, "Disables client generation")
		noServer          = set.Bool("no-server", false, "Disables server generation")
		// Debug options.
		debugIgnoreNotImplemented = set.String("debug.ignoreNotImplemented", "",
			"Ignore methods having functionality which is not implemented")
		debugNoerr = set.Bool("debug.noerr", false, "Ignore errors")

		// Logging options.
		logOptions ogenzap.Options

		// Profile options.
		cpuProfile     = set.String("cpuprofile", "", "Write cpu profile to file")
		memProfile     = set.String("memprofile", "", "Write memory profile to this file")
		memProfileRate = set.Int("memprofilerate", -1, "If > 0, sets runtime.MemProfileRate")

		// Version option.
		version = set.Bool("version", false, "Print version and exit")
	)
	logOptions.RegisterFlags(set)

	var (
		ctAliases gen.ContentTypeAliases

		filterPath    *regexp.Regexp
		filterMethods []string
	)
	set.Var(&ctAliases, "ct-alias", "Content type alias, e.g. text/x-markdown=text/plain")
	set.Func("filter-path", "Filter operations by path regex", func(s string) (err error) {
		filterPath, err = regexp.Compile(s)
		return err
	})
	set.Func("filter-methods", "Filter operations by HTTP methods (comma-separated)", func(s string) error {
		for _, m := range strings.Split(s, ",") {
			m = strings.TrimSpace(m)
			if m == "" {
				continue
			}
			filterMethods = append(filterMethods, m)
		}
		return nil
	})

	if err := set.Parse(os.Args[1:]); err != nil {
		return err
	}

	if *version {
		info, _ := ogenversion.GetInfo()
		fmt.Println(info)
		return nil
	}

	specPath := set.Arg(0)
	if set.NArg() == 0 || specPath == "" {
		set.Usage()
		return errors.New("no spec provided")
	}
	specPath = filepath.Clean(specPath)

	logger, err := ogenzap.Create(logOptions)
	if err != nil {
		return err
	}
	defer func() {
		_ = logger.Sync()
	}()

	if f := *cpuProfile; f != "" {
		f, err := os.Create(f)
		if err != nil {
			return errors.Wrap(err, "create cpu profile")
		}
		defer func() {
			_ = f.Close()
		}()

		if err := pprof.StartCPUProfile(f); err != nil {
			logger.Error("Start CPU profiling", zap.Error(err))
		} else {
			defer pprof.StopCPUProfile()
		}
	}
	if f := *memProfile; f != "" {
		f, err := os.Create(f)
		if err != nil {
			return errors.Wrap(err, "create memory profile")
		}
		defer func() {
			_ = f.Close()
		}()

		if *memProfileRate > 0 {
			runtime.MemProfileRate = *memProfileRate
		}
		defer func() {
			runtime.GC()
			if err := pprof.WriteHeapProfile(f); err != nil {
				logger.Error("Write memory profile", zap.Error(err))
			}
		}()
	}

	specDir, fileName := filepath.Split(specPath)
	opts := gen.Options{
		NoClient:             *noClient,
		NoServer:             *noServer,
		GenerateExampleTests: *generateTests,
		SkipTestRegex:        nil, // Set below.
		SkipUnimplemented:    *skipUnimplemented,
		InferSchemaType:      *inferTypes,
		AllowRemote:          *allowRemote,
		Remote: gen.RemoteOptions{
			ReadFile: func(p string) ([]byte, error) {
				return os.ReadFile(filepath.Join(specDir, p))
			},
		},
		Filters: gen.Filters{
			PathRegex: filterPath,
			Methods:   filterMethods,
		},
		IgnoreNotImplemented: strings.Split(*debugIgnoreNotImplemented, ","),
		ContentTypeAliases:   ctAliases,
		Filename:             fileName,
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

	data, err := os.ReadFile(specPath)
	if err != nil {
		return err
	}

	if err := generate(data, *packageName, *targetDir, *clean, opts); err != nil {
		if handleGenerateError(os.Stderr, fileName, data, err) {
			return errors.New("generation failed")
		}
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
