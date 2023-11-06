// Binary ogen generates go source code from OAS.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/go-faster/errors"
	"github.com/go-faster/yaml"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen"
	"github.com/ogen-go/ogen/gen/genfs"
	"github.com/ogen-go/ogen/internal/location"
	"github.com/ogen-go/ogen/internal/ogenversion"
	"github.com/ogen-go/ogen/internal/ogenzap"
)

func cleanDir(targetDir string, files []os.DirEntry) (rerr error) {
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
			// Do not stop on first error, try to remove all files.
			rerr = multierr.Append(rerr, err)
		}
	}
	return rerr
}

func generate(data []byte, packageName, targetDir string, clean bool, opts gen.Options) error {
	log := opts.Logger
	if log == nil {
		log = zap.NewNop()
	}

	spec, err := ogen.Parse(data)
	if err != nil {
		// For pretty error message, we need to pass location.File.
		return &location.Error{
			File: opts.Parser.File,
			Err:  errors.Wrap(err, "parse spec"),
		}
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

func handleGenerateError(w io.Writer, color bool, err error) (r bool) {
	defer func() {
		// Add trailing newline to the error message if error is handled.
		if r {
			_, _ = fmt.Fprintln(w)
		}
	}()

	if location.PrintPrettyError(w, color, err) {
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

func loadConfig(cfgPath string) (opts gen.Options, _ error) {
	if cfgPath == "" {
		for _, potentialPath := range []string{
			"ogen.yml",
			"ogen.yaml",
			".ogen.yml",
			".ogen.yaml",
		} {
			if _, err := os.Stat(potentialPath); err == nil {
				cfgPath = potentialPath
				goto read
			}
		}
		return opts, nil
	}
read:
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return opts, err
	}

	d := yaml.NewDecoder(bytes.NewReader(data))
	d.KnownFields(true)

	if err := d.Decode(&opts); err != nil {
		return opts, err
	}

	return opts, nil
}

func run() error {
	set := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	set.Usage = func() {
		_, toolName := filepath.Split(os.Args[0])
		_, _ = fmt.Fprintf(set.Output(), "Usage: %s [options] <spec>\n", toolName)
		set.PrintDefaults()
	}

	var (
		// Config flag.
		cfgPath = set.String("config", "", "Path to config file")

		// Generator options.
		targetDir   = set.String("target", "api", "Path to target dir")
		packageName = set.String("package", "api", "Target package name")
		clean       = set.Bool("clean", false, "Clean generated files before generation")

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

	opts, err := loadConfig(*cfgPath)
	if err != nil {
		return errors.Wrap(err, "load config")
	}
	opts.Logger = logger

	data, err := opts.SetLocation(specPath, gen.RemoteOptions{})
	if err != nil {
		return errors.Wrap(err, "resolve spec")
	}

	if err := generate(data, *packageName, *targetDir, *clean, opts); err != nil {
		if handleGenerateError(os.Stderr, logOptions.Color, err) {
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
