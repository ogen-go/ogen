package gen

import (
	"sort"
	"strings"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

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
		operations: nil,
		securities: map[string]*ir.Security{},
		tstorage:   newTStorage(),
		errType:    nil,
		router:     Router{},
		log:        opts.Logger,
	}

	if err := g.makeIR(api.Operations); err != nil {
		return nil, errors.Wrap(err, "make ir")
	}

	if err := g.route(); err != nil {
		return nil, &ErrBuildRouter{err: err}
	}

	return g, nil
}

func (g *Generator) makeIR(ops []*openapi.Operation) error {
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
			if err := g.fail(err); err != nil {
				return err
			}

			msg := err.Error()
			if uErr := unimplementedError(nil); errors.As(err, &uErr) {
				msg = uErr.Error()
			}
			log.Info("Skipping operation", zap.String("reason_error", msg))
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

	sort.SliceStable(g.operations, func(i, j int) bool {
		a, b := g.operations[i], g.operations[j]
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
