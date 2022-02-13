package gen

import (
	"regexp"
	"sort"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/internal/oas"
	"github.com/ogen-go/ogen/internal/oas/parser"
)

type Generator struct {
	opt        Options
	operations []*ir.Operation
	tstorage   *tstorage
	errType    *ir.StatusResponse
	router     Router
}

type Options struct {
	VerboseRoute         bool
	GenerateExampleTests bool
	SkipTestRegex        *regexp.Regexp
	InferSchemaType      bool
	Filters              Filters
	IgnoreNotImplemented []string
}

type Filters struct {
	PathRegex *regexp.Regexp
	Methods   []string
}

func (f Filters) accept(op *oas.Operation) bool {
	if f.PathRegex != nil && !f.PathRegex.MatchString(op.Path.String()) {
		return false
	}

	if len(f.Methods) > 0 {
		for _, m := range f.Methods {
			if strings.EqualFold(m, op.HTTPMethod) {
				return true
			}
		}
		return false
	}

	return true
}

func NewGenerator(spec *ogen.Spec, opts Options) (*Generator, error) {
	operations, err := parser.Parse(spec, opts.InferSchemaType)
	if err != nil {
		return nil, errors.Wrap(err, "parse")
	}

	g := &Generator{
		opt:      opts,
		tstorage: newTStorage(),
	}

	if err := g.makeIR(operations); err != nil {
		return nil, errors.Wrap(err, "make ir")
	}

	if err := g.route(); err != nil {
		return nil, errors.Wrap(err, "route")
	}

	return g, nil
}

func (g *Generator) makeIR(ops []*oas.Operation) error {
	if err := g.reduceDefault(ops); err != nil {
		return errors.Wrap(err, "reduce default")
	}

	for _, spec := range ops {
		if !g.opt.Filters.accept(spec) {
			continue
		}

		ctx := &genctx{
			path:   []string{"#", "paths", spec.Path.String(), spec.HTTPMethod},
			global: g.tstorage,
			local:  newTStorage(),
		}

		op, err := g.generateOperation(ctx, spec)
		if err != nil {
			err = errors.Wrapf(err, "path %q: %s",
				spec.Path.String(),
				strings.ToLower(spec.HTTPMethod),
			)
			if err := g.fail(err); err != nil {
				return err
			}

			continue
		}

		fixEqualRequests(ctx, op)
		fixEqualResponses(ctx, op)

		if err := g.tstorage.merge(ctx.local); err != nil {
			return err
		}

		g.operations = append(g.operations, op)
	}

	sort.SliceStable(g.operations, func(i, j int) bool {
		a, b := g.operations[i], g.operations[j]
		return strings.Compare(a.Name, b.Name) < 0
	})

	return nil
}
