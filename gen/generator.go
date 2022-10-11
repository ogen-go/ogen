package gen

import (
	"strings"

	"github.com/go-faster/errors"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
	"github.com/ogen-go/ogen/openapi/parser"
)

// Generator is OpenAPI-to-Go generator.
type Generator struct {
	opt        Options
	api        *openapi.API
	servers    []ir.Server
	operations []*ir.Operation
	securities map[string]*ir.Security
	tstorage   *tstorage
	errType    *ir.Response
	router     Router

	log *zap.Logger
}

// NewGenerator creates new Generator.
func NewGenerator(spec *ogen.Spec, opts Options) (*Generator, error) {
	opts.setDefaults()

	var external jsonschema.ExternalResolver
	if opts.AllowRemote {
		external = newExternalResolver(opts.Remote)
	}
	api, err := parser.Parse(spec, parser.Settings{
		External:   external,
		Filename:   opts.Filename,
		InferTypes: opts.InferSchemaType,
	})
	if err != nil {
		return nil, &ErrParseSpec{err: err}
	}

	g := &Generator{
		opt:        opts,
		api:        api,
		servers:    nil,
		operations: nil,
		securities: map[string]*ir.Security{},
		tstorage:   newTStorage(),
		errType:    nil,
		router:     Router{},
		log:        opts.Logger,
	}

	if err := g.makeIR(api); err != nil {
		return nil, errors.Wrap(err, "make ir")
	}

	if err := g.route(); err != nil {
		return nil, &ErrBuildRouter{err: err}
	}

	return g, nil
}

func (g *Generator) makeIR(api *openapi.API) error {
	if err := g.makeServers(api.Servers); err != nil {
		return errors.Wrap(err, "servers")
	}
	return g.makeOps(api.Operations)
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
		log := g.log.With(g.zapLocation(spec))

		if !g.opt.Filters.accept(spec) {
			log.Info("Skipping filtered operation")
			continue
		}

		ctx := &genctx{
			jsonptr: newJSONPointer("#", "paths", routePath, spec.HTTPMethod),
			global:  g.tstorage,
			local:   newTStorage(),
		}

		op, err := g.generateOperation(ctx, spec)
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

	slices.SortStableFunc(g.operations, func(a, b *ir.Operation) bool {
		return a.Name < b.Name
	})
	return nil
}

// Types returns generated types.
func (g *Generator) Types() map[string]*ir.Type {
	return g.tstorage.types
}

// API returns api schema.
func (g *Generator) API() *openapi.API {
	return g.api
}
