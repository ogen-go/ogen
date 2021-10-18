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
	}

	if err := w.Generate("params", "oas_params_gen.go", cfg); err != nil {
		return err
	}
	if err := w.Generate("param_decoders", "oas_param_dec_gen.go", cfg); err != nil {
		return err
	}
	if err := w.Generate("handlers", "oas_handlers_gen.go", cfg); err != nil {
		return err
	}
	if err := w.Generate("router", "oas_router_gen.go", cfg); err != nil {
		return err
	}
	if err := w.Generate("schemas", "oas_schemas_gen.go", cfg); err != nil {
		return err
	}
	if err := w.Generate("request_decoders", "oas_req_dec_gen.go", cfg); err != nil {
		return err
	}
	if err := w.Generate("request_encoders", "oas_req_enc_gen.go", cfg); err != nil {
		return err
	}
	if err := w.Generate("response_encoders", "oas_res_enc_gen.go", cfg); err != nil {
		return err
	}
	if err := w.Generate("response_decoders", "oas_res_dec_gen.go", cfg); err != nil {
		return err
	}
	if len(cfg.Interfaces) > 0 {
		if err := w.Generate("interfaces", "oas_iface_gen.go", cfg); err != nil {
			return err
		}
	}
	if err := w.Generate("validators", "oas_validators_gen.go", cfg); err != nil {
		return err
	}
	if err := w.Generate("schemas_json", "oas_json_schemas_gen.go", cfg); err != nil {
		return err
	}
	if err := w.Generate("server", "oas_server_gen.go", cfg); err != nil {
		return err
	}
	if err := w.Generate("client", "oas_client_gen.go", cfg); err != nil {
		return err
	}

	return nil
}
