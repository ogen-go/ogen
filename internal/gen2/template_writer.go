package gen

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/ogen-go/ogen/internal/ir"
)

type TemplateConfig struct {
	Package    string
	Operations []*ir.Operation
	Types      map[string]*ir.Type
	Interfaces map[string]*ir.Type
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
		Operations: g.operations,
		Types:      g.types,
		Interfaces: g.interfaces,
	}

	templates := []struct {
		name string
		file string
	}{
		{"schemas", "oas_schemas_gen.go"},
		{"interfaces", "oas_interfaces_gen.go"},
		{"params", "oas_params_gen.go"},
		{"param_decoders", "oas_param_dec_gen.go"},
		{"handlers", "oas_handlers_gen.go"},
		{"router", "oas_router_gen.go"},
		{"request_encoders", "oas_req_enc_gen.go"},
		{"request_decoders", "oas_req_dec_gen.go"},
		{"response_encoders", "oas_res_enc_gen.go"},
		{"response_decoders", "oas_res_dec_gen.go"},
		// {"validators", "oas_validators_gen.go"},
		// {"schemas_json", "oas_schemas_json_gen.go"},
		{"server", "oas_server_gen.go"},
		{"client", "oas_client_gen.go"},
	}

	for _, t := range templates {
		if err := w.Generate(t.name, t.file, cfg); err != nil {
			return err
		}
	}

	return nil
}
