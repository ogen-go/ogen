package gen

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/ogen-go/ogen/internal/ast"
)

type TemplateConfig struct {
	Package    string
	Methods    []*ast.Method
	Schemas    map[string]*ast.Schema
	Interfaces map[string]*ast.Interface
	Generics   []*ast.Schema
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
		return fmt.Errorf("name collision (already wrote %s)", fileName)
	}

	w.buf.Reset()
	if err := w.t.ExecuteTemplate(w.buf, templateName, cfg); err != nil {
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
		t:     vendoredTemplates(),
		buf:   new(bytes.Buffer),
		wrote: map[string]bool{},
	}

	cfg := TemplateConfig{
		Package:    pkgName,
		Schemas:    g.schemas,
		Methods:    g.methods,
		Interfaces: g.interfaces,
		Generics:   g.generics,
	}

	if err := w.Generate("params", "openapi_params_gen.go", cfg); err != nil {
		return err
	}
	if err := w.Generate("param_decoders", "openapi_param_decoders_gen.go", cfg); err != nil {
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
	if err := w.Generate("request_decoders", "openapi_request_decoders_gen.go", cfg); err != nil {
		return err
	}
	if err := w.Generate("request_encoders", "openapi_request_encoders_gen.go", cfg); err != nil {
		return err
	}
	if err := w.Generate("response_encoders", "openapi_response_encoders_gen.go", cfg); err != nil {
		return err
	}
	if err := w.Generate("response_decoders", "openapi_response_decoders_gen.go", cfg); err != nil {
		return err
	}
	if len(cfg.Interfaces) > 0 {
		if err := w.Generate("interfaces", "openapi_interfaces_gen.go", cfg); err != nil {
			return err
		}
	}
	if err := w.Generate("validators", "openapi_validators_gen.go", cfg); err != nil {
		return err
	}
	if err := w.Generate("schemas_json", "openapi_schemas_json_gen.go", cfg); err != nil {
		return err
	}
	if err := w.Generate("generic_json", "openapi_generic_json_gen.go", cfg); err != nil {
		return err
	}
	if err := w.Generate("server", "openapi_server_gen.go", cfg); err != nil {
		return err
	}
	if err := w.Generate("client", "openapi_client_gen.go", cfg); err != nil {
		return err
	}

	return nil
}
