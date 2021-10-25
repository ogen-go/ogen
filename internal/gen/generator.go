package gen

import (
	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/ast"
)

type Generator struct {
	opt           Options
	spec          *ogen.Spec
	methods       []*ast.Method
	schemas       map[string]*ast.Schema
	schemaRefs    map[string]*ast.Schema
	requestBodies map[string]*ast.RequestBody
	responses     map[string]*ast.Response
	interfaces    map[string]*ast.Interface
}

func (g *Generator) hasSchema(name string) bool {
	_, ok := g.schemas[name]
	return ok
}

func (g *Generator) freeSchemaName(names []string) (string, error) {
	for _, name := range names {
		if !g.hasSchema(name) {
			return name, nil
		}
	}
	return "", xerrors.Errorf("all of good names %v are taken", names)
}

type Options struct {
	SpecificMethodPath      string
	IgnoreUnspecifiedParams bool
	IgnoreNotImplemented    []string
}

func NewGenerator(spec *ogen.Spec, opts Options) (*Generator, error) {
	spec.Init()
	g := &Generator{
		opt:           opts,
		spec:          spec,
		schemas:       map[string]*ast.Schema{},
		schemaRefs:    map[string]*ast.Schema{},
		requestBodies: map[string]*ast.RequestBody{},
		responses:     map[string]*ast.Response{},
		interfaces:    map[string]*ast.Interface{},
	}

	if err := g.generateMethods(); err != nil {
		return nil, xerrors.Errorf("methods: %w", err)
	}

	g.simplify()
	if err := g.fix(); err != nil {
		return nil, xerrors.Errorf("fix: %w", err)
	}

	return g, nil
}
