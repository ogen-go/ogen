package gen

import (
	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/ast"
)

type Generator struct {
	opt     options
	spec    *ogen.Spec
	methods []*ast.Method

	schemas       map[string]*ast.Schema
	requestBodies map[string]*ast.RequestBody
	responses     map[string]*ast.Response
	interfaces    map[string]*ast.Interface
}

type options struct {
	// TODO: Remove
	debugIgnoreOptionals bool
	// TODO: Remove
	debugAllowEmptyObjects bool
	// TODO: Remove
	debugSkipUnspecified bool
}

type Option func(o *options)

// WithIgnoreOptionals ignores that optionals are not implemented.
func WithIgnoreOptionals(o *options) {
	o.debugIgnoreOptionals = true
}

// WithEmptyObjects allows empty objects.
func WithEmptyObjects(o *options) {
	o.debugAllowEmptyObjects = true
}

// WithSkipUnspecified skips unspecified types.
func WithSkipUnspecified(o *options) {
	o.debugSkipUnspecified = true
}

func NewGenerator(spec *ogen.Spec, opts ...Option) (*Generator, error) {
	o := options{}
	for _, f := range opts {
		f(&o)
	}

	initComponents(spec)
	g := &Generator{
		opt:           o,
		spec:          spec,
		schemas:       map[string]*ast.Schema{},
		requestBodies: map[string]*ast.RequestBody{},
		responses:     map[string]*ast.Response{},
		interfaces:    map[string]*ast.Interface{},
	}

	if err := g.generateMethods(); err != nil {
		return nil, xerrors.Errorf("methods: %w", err)
	}

	g.simplify()
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
