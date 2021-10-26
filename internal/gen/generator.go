package gen

import (
	"strings"

	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/internal/oas"
	"github.com/ogen-go/ogen/internal/oas/parser"
)

type Generator struct {
	opt        Options
	operations []*ir.Operation
	types      map[string]*ir.Type
	interfaces map[string]*ir.Type
	refs       map[string]*ir.Type
	wrapped    map[string]*ir.Type
}

type Options struct {
	SpecificMethodPath      string
	IgnoreUnspecifiedParams bool
	IgnoreNotImplemented    []string
}

func NewGenerator(spec *ogen.Spec, opts Options) (*Generator, error) {
	operations, err := parser.Parse(spec, parser.Options{
		IgnoreUnspecifiedParams: opts.IgnoreUnspecifiedParams,
	})
	if err != nil {
		return nil, err
	}

	g := &Generator{
		opt:        opts,
		types:      map[string]*ir.Type{},
		interfaces: map[string]*ir.Type{},
		refs:       map[string]*ir.Type{},
		wrapped:    map[string]*ir.Type{},
	}

	if err := g.makeIR(operations); err != nil {
		return nil, err
	}

	g.fix()
	for _, w := range g.wrapped {
		g.saveType(w)
	}
	return g, nil
}

func (g *Generator) makeIR(ops []*oas.Operation) error {
	for _, spec := range ops {
		op, err := g.generateOperation(spec)
		if err != nil {
			if g.shouldFail(err) {
				return xerrors.Errorf("%q: %s: %w",
					spec.Path(),
					strings.ToLower(spec.HTTPMethod),
					err,
				)
			}

			continue
		}

		g.operations = append(g.operations, op)
	}

	return nil
}
