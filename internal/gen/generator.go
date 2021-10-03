package gen

import (
	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen"
)

type Generator struct {
	opt     options
	spec    *ogen.Spec
	methods []*Method

	schemas       map[string]*Schema
	requestBodies map[string]*RequestBody
	responses     map[string]*Response
	interfaces    map[string]*Interface
}

type options struct {
	// TODO: Remove
	debugIgnoreOptionals bool
	// TODO: Remove
	debugIgnoreUnsupportedFormat bool
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

// WithIgnoreFormat ignores unsupported formats.
func WithIgnoreFormat(o *options) {
	o.debugIgnoreUnsupportedFormat = true
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
		schemas:       map[string]*Schema{},
		requestBodies: map[string]*RequestBody{},
		responses:     map[string]*Response{},
		interfaces:    map[string]*Interface{},
	}

	if err := g.generateComponents(); err != nil {
		return nil, xerrors.Errorf("components: %w", err)
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
