package gen

import (
	"bytes"
	"os"
	"text/template"

	"golang.org/x/xerrors"
)

// FileSystem represents a directory of generated package.
type FileSystem interface {
	WriteFile(baseName string, source []byte) error
}

type writer struct {
	pkg   string
	fs    FileSystem
	t     *template.Template
	buf   *bytes.Buffer
	wrote map[string]bool
}

// Generate executes template to file using config.
func (w *writer) Generate(templateName, fileName string, cfg config) error {
	if cfg.Package == "" {
		cfg.Package = w.pkg
	}
	if w.wrote[fileName] {
		return xerrors.Errorf("name collision (already wrote %s)", fileName)
	}

	w.buf.Reset()
	if err := w.t.ExecuteTemplate(w.buf, templateName, cfg); err != nil {
		return xerrors.Errorf("failed to execute template %s for %s: %w", templateName, fileName, err)
	}
	if err := w.fs.WriteFile(fileName, w.buf.Bytes()); err != nil {
		_ = os.WriteFile(fileName+".dump", w.buf.Bytes(), 0600)
		return xerrors.Errorf("failed to write file %s: %w", fileName, err)
	}
	w.wrote[fileName] = true

	return nil
}

// WriteSource writes generated definitions to fs.
func (g *Generator) WriteSource(fs FileSystem, pkgName string, t *template.Template) error {
	w := &writer{
		pkg:   pkgName,
		fs:    fs,
		t:     t,
		buf:   new(bytes.Buffer),
		wrote: map[string]bool{},
	}

	cfg := config{
		Components: g.components,
		Groups:     g.groups,
		Server:     g.server,
	}

	if err := w.Generate("components", "openapi_components.gen.go", cfg); err != nil {
		return err
	}
	if err := w.Generate("router", "openapi_router.gen.go", cfg); err != nil {
		return err
	}
	if err := w.Generate("server", "openapi_server.gen.go", cfg); err != nil {
		return err
	}

	return nil
}
