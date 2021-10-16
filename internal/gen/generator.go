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
	generics      []*ast.Schema
	schemas       map[string]*ast.Schema
	schemaRefs    map[string]*ast.Schema
	requestBodies map[string]*ast.RequestBody
	responses     map[string]*ast.Response
	interfaces    map[string]*ast.Interface
}

type Options struct {
	SpecificMethodPath      string
	IgnoreUnspecifiedParams bool
	IgnoreNotImplemented    bool
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

	g.generatePrimitives()
	g.simplify()
	g.fix()
	return g, nil
}

func (g *Generator) generatePrimitives() {
	for _, t := range []struct {
		Primitive string
		JSON      string
		ValueType string
	}{
		{Primitive: "string", JSON: "String", ValueType: "StringValue"},
		{Primitive: "int", JSON: "Int", ValueType: "NumberValue"},
		{Primitive: "int32", JSON: "Int32", ValueType: "NumberValue"},
		{Primitive: "int64", JSON: "Int64", ValueType: "NumberValue"},
		{Primitive: "float32", JSON: "Float32", ValueType: "NumberValue"},
		{Primitive: "float64", JSON: "Float64", ValueType: "NumberValue"},
		{Primitive: "bool", JSON: "Bool", ValueType: "BoolValue"},
	} {
		for _, v := range []struct {
			Optional bool
			Nil      bool
		}{
			{Optional: true, Nil: false},
			{Optional: false, Nil: true},
			{Optional: true, Nil: true},
		} {
			gt := &ast.Schema{
				Optional: v.Optional,
				Nil:      v.Nil,

				Kind:      ast.KindPrimitive,
				Primitive: t.Primitive,
				JSON: &ast.JSON{
					Read:      "Read" + t.JSON,
					Write:     "Write" + t.JSON,
					ValueType: t.ValueType,
				},
			}
			gt.Name = gt.GenericKind() + pascal(t.Primitive)
			g.generics = append(g.generics, gt)
		}
	}
}
