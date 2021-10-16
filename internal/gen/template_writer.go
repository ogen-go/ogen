package gen

import (
	"bytes"
	"fmt"
	"os"

	"github.com/open2b/scriggo"
	"github.com/open2b/scriggo/native"
)

// FileSystem represents a directory of generated package.
type FileSystem interface {
	WriteFile(baseName string, source []byte) error
}

type writer struct {
	fs    FileSystem
	buf   *bytes.Buffer
	wrote map[string]bool
}

// Generate executes template to file using config.
func (w *writer) Generate(templateName, fileName string, opts *scriggo.BuildOptions) error {
	if w.wrote[fileName] {
		return fmt.Errorf("name collision (already wrote %s)", fileName)
	}

	w.buf.Reset()
	template, err := scriggo.BuildTemplate(vendoredTemplates(), templateName, opts)
	if err != nil {
		return err
	}

	if err := template.Run(w.buf, nil, nil); err != nil {
		return fmt.Errorf("failed to execute template %s for %s: %w", templateName, fileName, err)
	}
	if err := w.fs.WriteFile(fileName, w.buf.Bytes()); err != nil {
		_ = os.WriteFile(fileName+".dump", w.buf.Bytes(), 0600)
		return fmt.Errorf("failed to write file %s: %w", fileName, err)
	}
	w.wrote[fileName] = true

	return nil
}

// WriteSource writes generated definitions to fs.
func (g *Generator) WriteSource(fs FileSystem, pkgName string) error {
	w := &writer{
		fs:    fs,
		buf:   new(bytes.Buffer),
		wrote: map[string]bool{},
	}

	globals := templateFuncs()
	globals["Package"] = pkgName
	globals["Schemas"] = &g.schemas
	globals["Methods"] = &g.methods
	globals["Interfaces"] = &g.interfaces

	opts := &scriggo.BuildOptions{
		Globals: globals,
		Packages: native.Packages{
			"ast": astPkg(),
		},
	}

	if err := w.Generate("schemas.tmpl", "openapi_schemas_gen.go", opts); err != nil {
		return err
	}
	return nil
}
