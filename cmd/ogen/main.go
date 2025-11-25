// Binary ogen generates go source code from OAS.
package main

import (
	"bytes"
	"cmp"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"slices"
	"strings"
	"time"

	"github.com/go-faster/errors"
	"github.com/go-faster/yaml"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen"
	"github.com/ogen-go/ogen/gen/genfs"
	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/internal/ogenversion"
	"github.com/ogen-go/ogen/internal/ogenzap"
	"github.com/ogen-go/ogen/location"
)

func cleanDir(targetDir string, files []os.DirEntry) (rerr error) {
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		if !strings.HasSuffix(name, "_gen.go") && !strings.HasSuffix(name, "_gen_test.go") {
			continue
		}
		if !strings.HasPrefix(name, "openapi") && !strings.HasPrefix(name, "oas") {
			continue
		}
		// Do not return error if file does not exist.
		if err := os.Remove(filepath.Join(targetDir, name)); err != nil && !os.IsNotExist(err) {
			// Do not stop on first error, try to remove all files.
			rerr = errors.Join(rerr, err)
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

	if msg, feature, ok := handleNotImplementedError(err); ok {
		_, _ = fmt.Fprintf(w, `
%s
Try to create ogen.yml with:

generator:
	ignore_not_implemented: [%q]

or

generator:
	ignore_not_implemented: ["all"]

to skip unsupported operations.
`, msg, feature)
		return true
	}

	return false
}

func handleNotImplementedError(err error) (msg, feature string, _ bool) {
	if notImplErr, ok := errors.Into[*gen.ErrNotImplemented](err); ok {
		msg := fmt.Sprintf("Feature %q is not implemented yet.\n", notImplErr.Name)
		return msg, notImplErr.Name, true
	}

	if ctErr, ok := errors.Into[*gen.ErrUnsupportedContentTypes](err); ok {
		if len(ctErr.ContentTypes) == 1 {
			msg = fmt.Sprintf(
				"Content type %q is unsupported.\n",
				ctErr.ContentTypes[0],
			)
		} else {
			msg = fmt.Sprintf(
				"Content types [%s] are unsupported.\n",
				strings.Join(ctErr.ContentTypes, ", "),
			)
		}
		return msg, "unsupported content types", true
	}

	if inferErr, ok := errors.Into[*gen.ErrFieldsDiscriminatorInference](err); ok {
		printTyp := func(sb *strings.Builder, typ *ir.Type) {
			if typ.Schema == nil {
				fmt.Fprintf(sb, "%q", typ.Name)
				return
			}
			if ref := typ.Schema.Ref; ref.IsZero() {
				fmt.Fprintf(sb, "%q", typ.Name)
			} else {
				fmt.Fprintf(sb, "%q", ref.Ptr)
			}
			ptr := typ.Schema.Pointer

			if pos, ok := ptr.Position(); ok {
				sb.WriteString(" (defined at ")
				at := pos.WithFilename(ptr.File().HumanName())
				sb.WriteString(at)
				sb.WriteString(")")
			}
		}
		var sb strings.Builder

		sb.WriteString("ogen failed to infer fields discriminator for type ")
		printTyp(&sb, inferErr.Sum)
		sb.WriteString(":\n")

		const (
			propertyLimit = 5 // 10
			usedByLimit   = 2
		)
		for _, bv := range inferErr.Types {
			sb.WriteString("\tvariant ")
			printTyp(&sb, bv.Type)
			sb.WriteString("\n")

			var (
				printedProperties int
				properties        = maps.Keys(bv.Fields)
			)
			// Sort by number of 'also used' types.
			//
			// It is likely to be properties to be fixed.
			slices.SortFunc(properties, func(a, b string) int {
				x, y := bv.Fields[a], bv.Fields[b]
				return cmp.Compare(len(x), len(y))
			})
			for _, field := range properties {
				if printedProperties >= propertyLimit {
					fmt.Fprintf(&sb, "\t\t...%d more properties...\n", len(properties))
					break
				}
				printedProperties++

				fmt.Fprintf(&sb, "\t\tproperty %q also used by\n", field)

				var (
					printedUsedBy int
					alsoUsedBy    = bv.Fields[field]
				)
				for _, typ := range alsoUsedBy {
					if printedUsedBy >= usedByLimit {
						fmt.Fprintf(&sb, "\t\t\t...%d more variants...\n", len(alsoUsedBy))
						break
					}
					printedUsedBy++

					sb.WriteString("\t\t\tvariant ")
					printTyp(&sb, typ)
					sb.WriteString("\n")
				}
			}
		}
		sb.WriteString("\n")
		return sb.String(), "discriminator inference", true
	}

	return msg, feature, false
}

func loadConfig(cfgPath string, log *zap.Logger) (opts gen.Options, _ error) {
	opts.Logger = log

	if cfgPath == "" {
		for _, potentialPath := range []string{
			"ogen.yml",
			"ogen.yaml",
			".ogen.yml",
			".ogen.yaml",
		} {
			if _, err := os.Stat(potentialPath); err == nil {
				cfgPath = potentialPath
				log.Debug("Found config file", zap.String("path", potentialPath))
				goto read
			}
		}
		log.Debug("No config file found")
		return opts, nil
	}
read:
	log.Debug("Reading config file", zap.String("path", cfgPath))
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

		// Parser options.
		strict = set.Bool("strict", false, "Disable cross-type constraint interpretation (reject pattern on numbers, min/max on strings)")

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

	opts, err := loadConfig(*cfgPath, logger)
	if err != nil {
		return errors.Wrap(err, "load config")
	}

	// Apply CLI flags that override config
	if *strict {
		strictVal := false
		opts.Parser.AllowCrossTypeConstraints = &strictVal
	}

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
