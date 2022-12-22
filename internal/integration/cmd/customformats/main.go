package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen"
	"github.com/ogen-go/ogen/gen/genfs"
	"github.com/ogen-go/ogen/internal/integration/customformats/hextype"
	"github.com/ogen-go/ogen/internal/integration/customformats/phonetype"
	"github.com/ogen-go/ogen/internal/integration/customformats/rgbatype"
	"github.com/ogen-go/ogen/internal/location"
	"github.com/ogen-go/ogen/internal/ogenzap"
	"github.com/ogen-go/ogen/internal/urlpath"
	"github.com/ogen-go/ogen/jsonschema"
)

func run(specPath, targetDir string) error {
	specPath, err := filepath.Abs(filepath.Clean(specPath))
	if err != nil {
		return err
	}

	data, err := os.ReadFile(specPath)
	if err != nil {
		return err
	}

	spec, err := ogen.Parse(data)
	if err != nil {
		return errors.Wrap(err, "parse spec")
	}

	u, err := urlpath.URLFromFilePath(specPath)
	if err != nil {
		return errors.Wrap(err, "convert file path to url")
	}

	l, err := ogenzap.Create(ogenzap.Options{Level: zap.DebugLevel})
	if err != nil {
		return errors.Wrap(err, "create logger")
	}
	defer func() {
		_ = l.Sync()
	}()

	_, fileName := filepath.Split(specPath)
	g, err := gen.NewGenerator(spec, gen.Options{
		RootURL: u,
		CustomFormats: gen.CustomFormatsMap{
			jsonschema.String: {
				"phone": phonetype.PhoneFormat,
				"rgba":  rgbatype.RGBAFormat,
				"hex":   hextype.HexFormat,
			},
		},
		File:   location.NewFile(fileName, specPath, data),
		Logger: l,
	})
	if err != nil {
		return errors.Wrap(err, "generate")
	}

	if err := os.MkdirAll(targetDir, 0o750); err != nil {
		return err
	}
	fs := genfs.FormattedSource{
		// FIXME(tdakkota): write source uses imports.Process which also uses go/format.
		// 	So, there is no reason to format source twice or provide a flag to disable formatting.
		Format: false,
		Root:   targetDir,
	}
	if err := g.WriteSource(fs, "api"); err != nil {
		return errors.Wrap(err, "write")
	}

	return nil
}

func main() {
	set := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	if err := set.Parse(os.Args[1:]); err != nil {
		panic(err)
	}

	specPath := set.Arg(0)
	if specPath == "" {
		panic("spec path is required")
	}

	targetDir := set.Arg(1)
	if targetDir == "" {
		targetDir = "api"
	}

	if err := run(specPath, targetDir); err != nil {
		panic(err)
	}
}
