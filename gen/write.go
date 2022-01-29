package gen

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"text/template"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/ir"
)

type TemplateConfig struct {
	Package    string
	Operations []*ir.Operation
	Types      map[string]*ir.Type
	URITypes   map[*ir.Type]struct{}
	Interfaces map[string]*ir.Type
	Error      *ir.StatusResponse
	ErrorType  *ir.Type
	Router     Router
}

// RegexStrings returns slice of all unique regex validators.
func (t TemplateConfig) RegexStrings() (r []string) {
	var (
		addRegex func(typ *ir.Type)
		m        = map[string]struct{}{}
		seen     = map[*ir.Type]struct{}{}
	)
	addRegex = func(typ *ir.Type) {
		_, skip := seen[typ]
		if typ == nil || skip {
			return
		}
		seen[typ] = struct{}{}

		if r := typ.Validators.String.Regex; r != nil {
			m[r.String()] = struct{}{}
		}
		for _, f := range typ.Fields {
			addRegex(f.Type)
		}
	}

	for _, typ := range t.Types {
		addRegex(typ)
	}
	for typ := range t.URITypes {
		addRegex(typ)
	}
	for _, typ := range t.Interfaces {
		addRegex(typ)
	}
	if t.Error != nil {
		addRegex(t.Error.NoContent)
		for _, typ := range t.Error.Contents {
			addRegex(typ)
		}
	}
	addRegex(t.ErrorType)

	for exp := range m {
		r = append(r, exp)
	}
	sort.Strings(r)
	return r
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
		return errors.Errorf("name collision (already wrote %s)", fileName)
	}

	w.buf.Reset()
	if err := w.t.ExecuteTemplate(w.buf, templateName, cfg); err != nil {
		return errors.Wrapf(err, "failed to execute template %s for %s", templateName, fileName)
	}
	if err := w.fs.WriteFile(fileName, w.buf.Bytes()); err != nil {
		_ = os.WriteFile(fileName+".dump", w.buf.Bytes(), 0600)
		return errors.Wrapf(err, "failed to write file %s", fileName)
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
		URITypes:   g.uriTypes,
		Interfaces: g.interfaces,
		Error:      g.errType,
		Router:     g.router,
	}
	if cfg.Error != nil {
		cfg.ErrorType = cfg.Error.Contents[ir.ContentTypeJSON]
	}
	for _, name := range []string{
		"schemas",
		"uri",
		"json",
		"interfaces",
		"param",
		"param_dec",
		"handlers",
		"req_enc",
		"req_dec",
		"res_enc",
		"res_dec",
		"validators",
		"server",
		"client",
		"cfg",
		"router",
	} {
		// Skip uri encode/decode if no types for that.
		if name == "uri" && len(g.uriTypes) == 0 {
			continue
		}

		fileName := fmt.Sprintf("oas_%s_gen.go", name)
		if err := w.Generate(name, fileName, cfg); err != nil {
			return errors.Wrapf(err, "%s", name)
		}
	}

	return nil
}
