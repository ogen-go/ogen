package gen

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"runtime/pprof"
	"slices"
	"sync"
	"text/template"

	"github.com/go-faster/errors"
	"golang.org/x/sync/errgroup"
	"golang.org/x/tools/imports"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/internal/xmaps"
	"github.com/ogen-go/ogen/ogenregex"
)

type TemplateConfig struct {
	Package           string
	Operations        []*ir.Operation
	DefaultOperations []*ir.Operation
	OperationGroups   []*ir.OperationGroup
	Webhooks          []*ir.Operation
	Types             map[string]*ir.Type
	Interfaces        map[string]*ir.Type
	Error             *ir.Response
	ErrorType         *ir.Type
	Servers           ir.Servers
	Securities        map[string]*ir.Security
	Router            Router
	WebhookRouter     WebhookRouter

	PathsClientEnabled        bool
	PathsServerEnabled        bool
	WebhookClientEnabled      bool
	WebhookServerEnabled      bool
	OpenTelemetryEnabled      bool
	SecurityReentrantEnabled  bool
	RequestValidationEnabled  bool
	ResponseValidationEnabled bool

	skipTestRegex *regexp.Regexp
}

// AnyClientEnabled returns true, if webhooks or paths client is enabled.
func (t TemplateConfig) AnyClientEnabled() bool {
	return t.PathsClientEnabled || t.WebhookClientEnabled
}

// AnyServerEnabled returns true, if webhooks or paths server is enabled.
func (t TemplateConfig) AnyServerEnabled() bool {
	return t.PathsServerEnabled || t.WebhookServerEnabled
}

// AnyInstrumentable returns true, if OpenTelemetry integration enabled and there is client/server to instrument.
func (t TemplateConfig) AnyInstrumentable() bool {
	return t.OpenTelemetryEnabled && (t.AnyClientEnabled() || t.AnyServerEnabled())
}

// ErrorGoType returns Go type of error.
func (t TemplateConfig) ErrorGoType() string {
	typ := t.ErrorType
	if typ.DoPassByPointer() {
		return "*" + typ.Go()
	}
	return typ.Go()
}

// SkipTest returns true, if test should be skipped.
func (t TemplateConfig) SkipTest(typ *ir.Type) bool {
	return t.skipTestRegex != nil && t.skipTestRegex.MatchString(typ.Name)
}

func (t TemplateConfig) collectStrings(cb func(typ *ir.Type) []string) []string {
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
		for _, media := range t.Error.Contents {
			add(media.Type)
		}
	}
	add(t.ErrorType)

	_ = walkOpTypes(t.Operations, func(t *ir.Type) error {
		add(t)
		return nil
	})
	_ = walkOpTypes(t.Webhooks, func(t *ir.Type) error {
		add(t)
		return nil
	})

	return xmaps.SortedKeys(m)
}

// RegexStrings returns slice of all unique regex validators.
func (t TemplateConfig) RegexStrings() []string {
	return t.collectStrings(func(typ *ir.Type) (r []string) {
		for _, exp := range []ogenregex.Regexp{
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
			// `RatString` return a string with integer value if denominator is 1.
			//
			// That makes string representation of `big.Rat` shorter and simpler.
			// Also, it is better for executable size.
			return []string{r.RatString()}
		}
		return nil
	})
}

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
			_ = os.WriteFile(fileName+".dump", generated, 0o644)
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

	features, err := g.opt.Features.Build()
	if err != nil {
		return errors.Wrap(err, "build feature set")
	}
	cfg := TemplateConfig{
		Package:                   pkgName,
		Operations:                g.operations,
		DefaultOperations:         g.defaultOperations,
		OperationGroups:           g.operationGroups,
		Webhooks:                  g.webhooks,
		Types:                     types,
		Interfaces:                interfaces,
		Error:                     g.errType,
		ErrorType:                 nil,
		Servers:                   g.servers,
		Securities:                g.securities,
		Router:                    g.router,
		WebhookRouter:             g.webhookRouter,
		PathsClientEnabled:        features.Has(PathsClient),
		PathsServerEnabled:        features.Has(PathsServer),
		WebhookClientEnabled:      features.Has(WebhooksClient) && len(g.webhooks) > 0,
		WebhookServerEnabled:      features.Has(WebhooksServer) && len(g.webhooks) > 0,
		OpenTelemetryEnabled:      features.Has(OgenOtel),
		SecurityReentrantEnabled:  features.Has(ClientSecurityReentrant),
		RequestValidationEnabled:  features.Has(ClientRequestValidation),
		ResponseValidationEnabled: features.Has(ServerResponseValidation),
		// Unused for now.
		skipTestRegex: nil,
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
	var (
		genClient = cfg.AnyClientEnabled()
		genServer = cfg.AnyServerEnabled()
	)
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
		{"servers", len(g.servers) > 0},
		{"router", genServer},
		{"defaults", g.hasDefaultFields()},
		{"security", (genClient || genServer) && len(g.securities) > 0},
		{"test_examples", features.Has(DebugExampleTests)},
		{"faker", features.Has(DebugExampleTests)},
		{"unimplemented", features.Has(OgenUnimplemented) && genServer},
		{"labeler", features.Has(OgenOtel) && genServer},
		{"operations", (genClient || genServer)},
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
