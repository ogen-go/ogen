package gen

import (
	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/ast"
)

type Generator struct {
	opt     Options
	spec    *ogen.Spec
	methods []*ast.Method

	schemas       map[string]*ast.Schema
	schemaRefs    map[string]*ast.Schema
	requestBodies map[string]*ast.RequestBody
	responses     map[string]*ast.Response
	interfaces    map[string]*ast.Interface
}

type Options struct {
	SpecificMethodPath      string
	IgnoreUnspecifiedParams bool
	IgnoreUnsupportedParams bool
	IgnoreOneOf             bool
	IgnoreAnyOf             bool
	IgnoreAllOf             bool
}

func NewGenerator(spec *ogen.Spec, opts Options) (*Generator, error) {
	initComponents(spec)
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
	g.fix()
	return g, nil
}

func initComponents(spec *ogen.Spec) {
	if spec.Components == nil {
		spec.Components = &ogen.Components{}
	}

	c := spec.Components
	if c.Schemas == nil {
		c.Schemas = make(map[string]ogen.Schema)
	}
	if c.Responses == nil {
		c.Responses = make(map[string]ogen.Response)
	}
	if c.Parameters == nil {
		c.Parameters = make(map[string]ogen.Parameter)
	}
	if c.RequestBodies == nil {
		c.RequestBodies = make(map[string]ogen.RequestBody)
	}
}
