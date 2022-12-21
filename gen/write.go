package gen

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"text/template"

	"github.com/go-faster/errors"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"
	"golang.org/x/tools/imports"

	"github.com/ogen-go/ogen/gen/ir"
)

// FileSystem represents a directory of generated package.
type FileSystem interface {
	WriteFile(baseName string, source []byte) error
}

type writer struct {
	fs FileSystem
	t  *template.Template
}

// generatorBufSize is 1 MB, it's enough for most mid-size specs.
const generatorBufSize = 1024 * 1024

var bufPool = sync.Pool{
	New: func() interface{} {
		var b bytes.Buffer
		b.Grow(generatorBufSize)
		b.Reset()
		return &b
	},
}

func getBuffer() *bytes.Buffer {
	b := bufPool.Get().(*bytes.Buffer)
	b.Reset()
	return b
}

func putBuffer(b *bytes.Buffer) {
	if b.Cap() > generatorBufSize {
		return
	}
	bufPool.Put(b)
}

// Generate executes template to file using config.
func (w *writer) Generate(templateName, fileName string, cfg TemplateConfig) (rerr error) {
	buf := getBuffer()
	defer putBuffer(buf)

	if err := w.t.ExecuteTemplate(buf, templateName, cfg); err != nil {
		return errors.Wrap(err, "execute")
	}

	generated := buf.Bytes()
	defer func() {
		if rerr != nil {
			_ = os.WriteFile(fileName+".dump", generated, 0o600)
		}
	}()

	formatted, err := imports.Process(fileName, generated, nil)
	if err != nil {
		return &ErrGoFormat{
			err: err,
		}
	}

	if err := w.fs.WriteFile(fileName, formatted); err != nil {
		return errors.Wrap(err, "write")
	}

	return nil
}

// WriteSource writes generated definitions to fs.
func (g *Generator) WriteSource(fs FileSystem, pkgName string) error {
	w := &writer{
		fs: fs,
		t:  vendoredTemplates(),
	}

	// Historically we separate interfaces from other types.
	// This is done for backward compatibility.
	types := make(map[string]*ir.Type, len(g.tstorage.types))
	interfaces := make(map[string]*ir.Type)
	for name, t := range g.tstorage.types {
		if t.IsInterface() {
			interfaces[name] = t
			continue
		}

		types[name] = t
	}

	cfg := TemplateConfig{
		Package:       pkgName,
		Operations:    g.operations,
		Webhooks:      g.webhooks,
		Types:         types,
		Interfaces:    interfaces,
		Error:         g.errType,
		ErrorType:     nil,
		Servers:       g.servers,
		Securities:    g.securities,
		Router:        g.router,
		WebhookRouter: g.webhookRouter,
		ClientEnabled: !g.opt.NoClient,
		ServerEnabled: !g.opt.NoServer,
		skipTestRegex: g.opt.SkipTestRegex,
	}
	if cfg.Error != nil {
		if len(cfg.Error.Contents) != 1 {
			panic(unreachable("error type must have exactly one content type"))
		}
		for _, media := range cfg.Error.Contents {
			if media.Encoding.JSON() {
				cfg.ErrorType = media.Type
				break
			}
		}
	}

	grp, ctx := errgroup.WithContext(context.Background())
	grp.SetLimit(runtime.GOMAXPROCS(0))
	generate := func(fileName, templateName string) {
		grp.Go(func() (err error) {
			labels := pprof.Labels("template", templateName)
			pprof.Do(ctx, labels, func(ctx context.Context) {
				err = w.Generate(templateName, fileName, cfg)
			})
			if err != nil {
				return errors.Wrapf(err, "template %q", templateName)
			}
			return nil
		})
	}
	genClient, genServer := !g.opt.NoClient, !g.opt.NoServer
	for _, t := range []struct {
		name    string
		enabled bool
	}{
		{"schemas", true},
		{"uri", g.hasURIObjectParams()},
		{"json", g.hasJSON()},
		{"interfaces", (genClient || genServer) && len(interfaces) > 0},
		{"parameters", g.hasParams()},
		{"handlers", genServer},
		{"request_encoders", genClient},
		{"request_decoders", genServer},
		{"response_encoders", genServer},
		{"response_decoders", genClient},
		{"validators", g.hasValidators()},
		{"middleware", genServer},
		{"server", genServer},
		{"client", genClient},
		{"cfg", true},
		{"ogenreflect", genClient || genServer},
		{"servers", len(g.servers) > 0},
		{"router", genServer},
		{"defaults", g.hasDefaultFields()},
		{"security", (genClient || genServer) && len(g.securities) > 0},
		{"test_examples", g.opt.GenerateExampleTests},
		{"faker", g.opt.GenerateExampleTests},
		{"unimplemented", !g.opt.SkipUnimplemented && genServer},
	} {
		t := t
		if !t.enabled {
			continue
		}

		fileName := fmt.Sprintf("oas_%s_gen.go", t.name)
		if t.name == "test_examples" {
			fileName = fmt.Sprintf("oas_%s_gen_test.go", t.name)
		}

		generate(fileName, t.name)
	}

	return grp.Wait()
}

func (g *Generator) hasAnyType(check func(t *ir.Type) bool) bool {
	for _, t := range g.tstorage.types {
		if check(t) {
			return true
		}
	}
	return false
}

func (g *Generator) hasDefaultFields() bool {
	return g.hasAnyType((*ir.Type).HasDefaultFields)
}

func (g *Generator) hasJSON() bool {
	return g.hasAnyType(func(t *ir.Type) bool {
		return t.HasFeature("json")
	})
}

func (g *Generator) hasValidators() bool {
	return g.hasAnyType((*ir.Type).NeedValidation)
}

func (g *Generator) hasParams() bool {
	hasParams := func(op *ir.Operation) bool {
		return len(op.Params) > 0
	}
	return slices.ContainsFunc(g.operations, hasParams) ||
		slices.ContainsFunc(g.webhooks, hasParams)
}

func (g *Generator) hasURIObjectParams() bool {
	return g.hasAnyType(func(t *ir.Type) bool {
		return t.IsStruct() && t.HasFeature("uri")
	})
}
