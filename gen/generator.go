package gen

import (
	"fmt"
	"go/token"
	"reflect"
	"strings"

	"github.com/go-faster/errors"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/internal/xmaps"
	"github.com/ogen-go/ogen/internal/xslices"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
	"github.com/ogen-go/ogen/openapi/parser"
)

// Generator is OpenAPI-to-Go generator.
type Generator struct {
	opt           Options
	api           *openapi.API
	servers       []ir.Server
	operations    []*ir.Operation
	webhooks      []*ir.Operation
	securities    map[string]*ir.Security
	tstorage      *tstorage
	errType       *ir.Response
	webhookRouter WebhookRouter
	router        Router

	customFormats map[jsonschema.SchemaType]map[string]ir.CustomFormat
	imports       []string

	log *zap.Logger
}

// NewGenerator creates new Generator.
func NewGenerator(spec *ogen.Spec, opts Options) (*Generator, error) {
	opts.setDefaults()

	var external jsonschema.ExternalResolver
	if opts.AllowRemote {
		external = jsonschema.NewExternalResolver(opts.Remote)
	}
	api, err := parser.Parse(spec, parser.Settings{
		External:   external,
		File:       opts.File,
		RootURL:    opts.RootURL,
		InferTypes: opts.InferSchemaType,
	})
	if err != nil {
		return nil, &ErrParseSpec{err: err}
	}

	g := &Generator{
		opt:           opts,
		api:           api,
		servers:       nil,
		operations:    nil,
		webhooks:      nil,
		securities:    map[string]*ir.Security{},
		tstorage:      newTStorage(),
		errType:       nil,
		webhookRouter: WebhookRouter{},
		router:        Router{},
		customFormats: map[jsonschema.SchemaType]map[string]ir.CustomFormat{},
		log:           opts.Logger,
	}

	if err := g.makeCustomFormats(); err != nil {
		return nil, errors.Wrap(err, "make custom formats")
	}

	if err := g.makeIR(api); err != nil {
		return nil, errors.Wrap(err, "make ir")
	}

	if err := g.route(); err != nil {
		return nil, &ErrBuildRouter{err: err}
	}

	return g, nil
}

func (g *Generator) makeCustomFormats() error {
	importPaths := map[string]string{}

	makeExternal := func(typ reflect.Type) (ir.ExternalType, error) {
		path := typ.PkgPath()
		if path == "main" {
			return ir.ExternalType{}, errors.New("type must be in importable package")
		}
		if n := typ.Name(); n == "" || !token.IsExported(n) {
			return ir.ExternalType{}, errors.New("type must be named and exported")
		}

		importName, ok := importPaths[path]
		if !ok {
			importName = fmt.Sprintf("custom%d", len(importPaths))
			importPaths[path] = importName
			g.imports = append(g.imports, fmt.Sprintf("%s %q", importName, path))
		}

		return ir.ExternalType{
			Pkg:  importName,
			Type: typ,
		}, nil
	}

	for _, jsonTyp := range xmaps.SortedKeys(g.opt.CustomFormats) {
		formats := g.opt.CustomFormats[jsonTyp]
		for _, format := range xmaps.SortedKeys(formats) {
			def := formats[format]

			if _, ok := g.customFormats[jsonTyp]; !ok {
				g.customFormats[jsonTyp] = map[string]ir.CustomFormat{}
			}

			f, err := func() (f ir.CustomFormat, _ error) {
				goName, err := pascalNonEmpty(format)
				if err != nil {
					return f, errors.Wrap(err, "generate go name")
				}

				typ, err := makeExternal(def.typ)
				if err != nil {
					return f, errors.Wrap(err, "format type")
				}

				json, err := makeExternal(def.json)
				if err != nil {
					return f, errors.Wrap(err, "json encoding")
				}

				text, err := makeExternal(def.text)
				if err != nil {
					return f, errors.Wrap(err, "text encoding")
				}

				return ir.CustomFormat{
					Name:   format,
					GoName: goName,
					Type:   typ,
					JSON:   json,
					Text:   text,
				}, nil
			}()
			if err != nil {
				return errors.Wrapf(err, "custom format %q:%q", jsonTyp, format)
			}

			g.customFormats[jsonTyp][format] = f
		}
	}

	return nil
}

func (g *Generator) makeIR(api *openapi.API) error {
	if err := g.makeServers(api.Servers); err != nil {
		return errors.Wrap(err, "servers")
	}
	if err := g.makeWebhooks(api.Webhooks); err != nil {
		return errors.Wrap(err, "webhooks")
	}
	if err := g.makeOps(api.Operations); err != nil {
		return errors.Wrap(err, "operations")
	}
	return nil
}

func (g *Generator) makeServers(servers []openapi.Server) error {
	for _, s := range servers {
		// Ignore servers without name.
		if s.Name == "" {
			continue
		}
		server, err := g.generateServer(s)
		if err != nil {
			return errors.Wrapf(err, "generate server %q", s.Name)
		}
		g.servers = append(g.servers, server)
	}
	return nil
}

func (g *Generator) makeOps(ops []*openapi.Operation) error {
	if err := g.reduceDefault(ops); err != nil {
		return errors.Wrap(err, "reduce default")
	}

	for _, spec := range ops {
		routePath := spec.Path.String()
		log := g.log.With(zapPosition(spec))

		if !g.opt.Filters.accept(spec) {
			log.Info("Skipping filtered operation")
			continue
		}

		ctx := &genctx{
			global: g.tstorage,
			local:  newTStorage(),
		}

		op, err := g.generateOperation(ctx, "", spec)
		if err != nil {
			err = errors.Wrapf(err, "path %q: %s",
				routePath,
				strings.ToLower(spec.HTTPMethod),
			)
			if err := g.trySkip(err, "Skipping operation", spec); err != nil {
				return err
			}
			continue
		}

		if err := fixEqualRequests(ctx, op); err != nil {
			return errors.Wrap(err, "fix requests")
		}
		if err := fixEqualResponses(ctx, op); err != nil {
			return errors.Wrap(err, "fix responses")
		}

		if err := g.tstorage.merge(ctx.local); err != nil {
			return err
		}

		g.operations = append(g.operations, op)
	}
	sortOperations(g.operations)
	return nil
}

func (g *Generator) makeWebhooks(webhooks []openapi.Webhook) error {
	for _, w := range webhooks {
		if w.Name == "" {
			rerr := errors.New("webhook name is empty")
			if err := g.trySkip(rerr, "Skipping webhook", w); err != nil {
				return err
			}
			continue
		}
		if len(w.Operations) == 0 {
			continue
		}

		var whinfo = &ir.WebhookInfo{
			Name: w.Name,
		}
		for _, spec := range w.Operations {
			log := g.log.With(zapPosition(spec))

			if !g.opt.Filters.accept(spec) {
				log.Info("Skipping filtered operation")
				continue
			}

			xslices.Filter(&spec.Parameters, func(p *openapi.Parameter) bool {
				if p.In.Path() {
					log.Warn("Webhooks can't have path parameters",
						zap.String("name", p.Name),
						zap.String("in", string(p.In)),
					)
					return false
				}
				return true
			})

			ctx := &genctx{
				global: g.tstorage,
				local:  newTStorage(),
			}

			op, err := g.generateOperation(ctx, w.Name, spec)
			if err != nil {
				err = errors.Wrapf(err, "webhook %q: %s",
					w.Name,
					strings.ToLower(spec.HTTPMethod),
				)
				if err := g.trySkip(err, "Skipping operation", spec); err != nil {
					return err
				}
				continue
			}
			op.WebhookInfo = whinfo

			if err := fixEqualRequests(ctx, op); err != nil {
				return errors.Wrap(err, "fix requests")
			}
			if err := fixEqualResponses(ctx, op); err != nil {
				return errors.Wrap(err, "fix responses")
			}

			if err := g.tstorage.merge(ctx.local); err != nil {
				return err
			}

			g.webhooks = append(g.webhooks, op)
		}
	}
	sortOperations(g.webhooks)
	return nil
}

func sortOperations(ops []*ir.Operation) {
	slices.SortStableFunc(ops, func(a, b *ir.Operation) bool {
		return a.Name < b.Name
	})
}

// Types returns generated types.
func (g *Generator) Types() map[string]*ir.Type {
	return g.tstorage.types
}

// API returns api schema.
func (g *Generator) API() *openapi.API {
	return g.api
}
