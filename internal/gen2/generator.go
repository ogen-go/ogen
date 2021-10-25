package gen

import (
	"github.com/ogen-go/ogen"
	ast "github.com/ogen-go/ogen/internal/ast2"
	"github.com/ogen-go/ogen/internal/ast2/parser"
	"github.com/ogen-go/ogen/internal/ir"
	"golang.org/x/xerrors"
)

type Generator struct {
	methods []*ir.Method
	types   map[string]*ir.Type
	refs    map[string]*ir.Type
}

func New(spec *ogen.Spec) (*Generator, error) {
	methods, err := parser.Parse(spec)
	if err != nil {
		return nil, err
	}

	g := &Generator{
		types: map[string]*ir.Type{},
		refs:  map[string]*ir.Type{},
	}

	if err := g.makeIR(methods); err != nil {
		return nil, err
	}

	g.fix()
	return g, nil
}

func (g *Generator) makeIR(methods []*ast.Method) error {
	for _, spec := range methods {
		m, err := g.generateMethod(spec)
		if err != nil {
			return xerrors.Errorf("'%s': %s: %w", spec.Path(), spec.HTTPMethod, err)
		}

		g.methods = append(g.methods, m)
	}

	return nil
}
