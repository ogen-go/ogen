package gen

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"sort"
	"text/template"

	"github.com/go-faster/errors"
	"golang.org/x/tools/imports"

	"github.com/ogen-go/ogen/gen/ir"
)

type TemplateConfig struct {
	Package       string
	Operations    []*ir.Operation
	Types         map[string]*ir.Type
	Interfaces    map[string]*ir.Type
	Error         *ir.Response
	ErrorType     *ir.Type
	Securities    map[string]*ir.Security
	Router        Router
	ClientEnabled bool
	ServerEnabled bool

	skipTestRegex *regexp.Regexp
}

// SkipTest returns true, if test should be skipped.
func (t TemplateConfig) SkipTest(typ *ir.Type) bool {
	return t.skipTestRegex != nil && t.skipTestRegex.MatchString(typ.Name)
}

func (t TemplateConfig) collectStrings(cb func(typ *ir.Type) []string) (r []string) {
	var (
		add  func(typ *ir.Type)
		m    = map[string]struct{}{}
		seen = map[*ir.Type]struct{}{}
	)
	add = func(typ *ir.Type) {
		_, skip := seen[typ]
		if typ == nil || skip {
			return
		}
		seen[typ] = struct{}{}
		for _, got := range cb(typ) {
			m[got] = struct{}{}
		}

		for _, f := range typ.Fields {
			add(f.Type)
		}
		for _, f := range typ.SumOf {
			add(f)
		}
		add(typ.AliasTo)
		add(typ.PointerTo)
		add(typ.GenericOf)
		add(typ.Item)
	}

	for _, typ := range t.Types {
		add(typ)
	}
	for _, typ := range t.Interfaces {
		add(typ)
	}
	if t.Error != nil {
		add(t.Error.NoContent)
		for _, typ := range t.Error.Contents {
			add(typ)
		}
	}
	add(t.ErrorType)

	for exp := range m {
		r = append(r, exp)
	}
	sort.Strings(r)
	return r
}

// RegexStrings returns slice of all unique regex validators.
func (t TemplateConfig) RegexStrings() []string {
	return t.collectStrings(func(typ *ir.Type) (r []string) {
		for _, exp := range []*regexp.Regexp{
			typ.Validators.String.Regex,
			typ.MapPattern,
		} {
			if exp == nil {
				continue
			}
			r = append(r, exp.String())
		}
		return r
	})
}

// RatStrings returns slice of all unique big.Rat (multipleOf validation).
func (t TemplateConfig) RatStrings() []string {
	return t.collectStrings(func(typ *ir.Type) []string {
		if r := typ.Validators.Float.MultipleOf; r != nil {
			return []string{r.String()}
		}
		return nil
	})
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

	b, err := imports.Process(fileName, w.buf.Bytes(), nil)
	if err != nil {
		_ = os.WriteFile(fileName+".dump", w.buf.Bytes(), 0o600)
		return errors.Wrap(err, "format imports")
	}

	if err := w.fs.WriteFile(fileName, b); err != nil {
		_ = os.WriteFile(fileName+".dump", b, 0o600)
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

	// Historically we separate interfaces from other types.
	// This is done for backward compatibility.
	interfaces := make(map[string]*ir.Type)
	for name, t := range g.tstorage.types {
		if t.IsInterface() {
			delete(g.tstorage.types, name)
			interfaces[name] = t
		}
	}

	cfg := TemplateConfig{
		Package:       pkgName,
		Operations:    g.operations,
		Types:         g.tstorage.types,
		Interfaces:    interfaces,
		Error:         g.errType,
		ErrorType:     nil,
		Securities:    g.securities,
		Router:        g.router,
		ClientEnabled: !g.opt.NoClient,
		ServerEnabled: !g.opt.NoServer,
		skipTestRegex: g.opt.SkipTestRegex,
	}
	if cfg.Error != nil {
		cfg.ErrorType = cfg.Error.Contents[ir.ContentTypeJSON]
	}

	genClient, genServer := !g.opt.NoClient, !g.opt.NoServer
	for _, t := range []struct {
		name    string
		enabled bool
	}{
		{"schemas", true},
		{"uri", g.hasURIObjectParams()},
		{"json", true},
		{"interfaces", genClient || genServer},
		{"parameters", true},
		{"handlers", genServer},
		{"request_encoders", genClient},
		{"request_decoders", genServer},
		{"response_encoders", genServer},
		{"response_decoders", genClient},
		{"validators", true},
		{"server", genServer},
		{"client", genClient},
		{"cfg", true},
		{"router", genServer},
		{"defaults", true},
		{"security", genClient || genServer},
		{"test_examples", g.opt.GenerateExampleTests},
		{"faker", g.opt.GenerateExampleTests},
		{"unimplemented", !g.opt.SkipUnimplemented && genServer},
	} {
		if !t.enabled {
			continue
		}

		fileName := fmt.Sprintf("oas_%s_gen.go", t.name)
		if t.name == "test_examples" {
			fileName = fmt.Sprintf("oas_%s_gen_test.go", t.name)
		}

		if err := w.Generate(t.name, fileName, cfg); err != nil {
			return errors.Wrapf(err, "%s", t.name)
		}
	}

	return nil
}

func (g *Generator) hasURIObjectParams() bool {
	for _, t := range g.tstorage.types {
		if t.HasFeature("uri") && t.Is(ir.KindStruct) {
			return true
		}
	}
	return false
}
