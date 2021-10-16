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
	generics      []*ast.Generic
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
	}{
		{Primitive: "string", JSON: "String"},
		{Primitive: "int", JSON: "Int"},
		{Primitive: "int32", JSON: "Int32"},
		{Primitive: "int64", JSON: "Int64"},
		{Primitive: "float32", JSON: "Float32"},
		{Primitive: "float64", JSON: "Float64"},
		{Primitive: "bool", JSON: "Bool"},
	} {
		for _, v := range []struct {
			Optional bool
			Nil      bool
		}{
			{Optional: true, Nil: false},
			{Optional: false, Nil: true},
			{Optional: true, Nil: true},
		} {
			gt := &ast.Generic{
				Optional: v.Optional,
				Nil:      v.Nil,

				Schema: ast.Schema{
					Kind:      ast.KindPrimitive,
					Primitive: t.Primitive,
					JSON: &ast.JSON{
						Read:  "Read" + t.JSON,
						Write: "Write" + t.JSON,
					},
				},
			}
			gt.Name = gt.GenericKind() + pascal(t.Primitive)
			g.generics = append(g.generics, gt)
		}
	}
}
