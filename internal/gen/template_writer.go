package gen

import (
	"bytes"
	"os"
	"text/template"

	"golang.org/x/xerrors"
)

type TemplateConfig struct {
	Package    string
	Methods    []*Method
	Schemas    map[string]*Schema
	Interfaces map[string]*Interface
}

// FileSystem represents a directory of generated package.
type FileSystem interface {
	WriteFile(baseName string, source []byte) error
}

type writer struct {
	fs    FileSystem
	t     *template.Template
	buf   *bytes.Buffer
	wrote map[string]bool
}

// Generate executes template to file using config.
func (w *writer) Generate(templateName, fileName string, cfg TemplateConfig) error {
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
func (g *Generator) WriteSource(fs FileSystem, pkgName string) error {
	w := &writer{
		fs:    fs,
		t:     vendoredTemplates(),
		buf:   new(bytes.Buffer),
		wrote: map[string]bool{},
	}

	cfg := TemplateConfig{
		Package:    pkgName,
		Schemas:    g.schemas,
		Methods:    g.methods,
		Interfaces: g.interfaces,
	}

	if err := w.Generate("parameters", "openapi_parameters_gen.go", cfg); err != nil {
		return err
	}
	if err := w.Generate("handlers", "openapi_handlers_gen.go", cfg); err != nil {
		return err
	}
	if err := w.Generate("router", "openapi_router_gen.go", cfg); err != nil {
		return err
	}
	if err := w.Generate("schemas", "openapi_schemas_gen.go", cfg); err != nil {
		return err
	}
	if err := w.Generate("requests", "openapi_requests_gen.go", cfg); err != nil {
		return err
	}
	if err := w.Generate("responses", "openapi_responses_gen.go", cfg); err != nil {
		return err
	}
	if len(cfg.Interfaces) > 0 {
		if err := w.Generate("interfaces", "openapi_interfaces_gen.go", cfg); err != nil {
			return err
		}
	}
	if err := w.Generate("server", "openapi_server_gen.go", cfg); err != nil {
		return err
	}

	return nil
}
