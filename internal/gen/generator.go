package gen

import (
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
	types      map[string]*ir.Type
	interfaces map[string]*ir.TypeInterface
	refs       struct {
		schemas   map[string]*ir.Type
		responses map[string]*ir.StatusResponse
	}
	wrapped struct {
		types     map[string]*ir.Type
		responses map[string]*ir.StatusResponse
	}
	uritypes map[*ir.Type]struct{}
}

type Options struct {
	SpecificMethodPath   string
	IgnoreNotImplemented []string
}

func NewGenerator(spec *ogen.Spec, opts Options) (*Generator, error) {
	operations, err := parser.Parse(spec)
	if err != nil {
		return nil, err
	}

	g := &Generator{
		opt:        opts,
		types:      map[string]*ir.Type{},
		interfaces: map[string]*ir.TypeInterface{},
		refs: struct {
			schemas   map[string]*ir.Type
			responses map[string]*ir.StatusResponse
		}{
			schemas:   map[string]*ir.Type{},
			responses: map[string]*ir.StatusResponse{},
		},
		wrapped: struct {
			types     map[string]*ir.Type
			responses map[string]*ir.StatusResponse
		}{
			types:     map[string]*ir.Type{},
			responses: map[string]*ir.StatusResponse{},
		},
		uritypes: map[*ir.Type]struct{}{},
	}

	if err := g.makeIR(operations); err != nil {
		return nil, err
	}
	for _, w := range g.wrapped.types {
		g.saveType(w)
	}
	g.fix()
	g.wrapGenerics()
	return g, nil
}

func (g *Generator) makeIR(ops []*oas.Operation) error {
	for _, spec := range ops {
		op, err := g.generateOperation(spec)
		if err != nil {
			if g.shouldFail(err) {
				return errors.Wrapf(err, "%q: %s",
					spec.Path(),
					strings.ToLower(spec.HTTPMethod),
				)
			}

			continue
		}

		g.operations = append(g.operations, op)
	}

	sort.SliceStable(g.operations, func(i, j int) bool {
		a, b := g.operations[i], g.operations[j]
		return strings.Compare(a.Name, b.Name) < 0
	})

	return nil
}
