package gen

import (
	"github.com/ogen-go/ogen/internal/ast"
)

func (g *Generator) generatePrimitives() {
	for _, name := range []string{
		"string",
		"int",
		"int32",
		"int64",
		"float32",
		"float64",
		"bool",
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
				Optional:  v.Optional,
				Nil:       v.Nil,
				Kind:      ast.KindPrimitive,
				Primitive: name,
			}
			gt.Name = gt.GenericKind() + pascal(name)
			g.generics = append(g.generics, gt)
		}
	}
}
