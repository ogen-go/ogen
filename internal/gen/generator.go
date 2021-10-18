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

	g.generatePrimitiveGenerics()
	g.simplify()
	g.fix()
	return g, nil
}
