package gen

import (
	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/ast"
)

func (g *Generator) resolveSchema(ref string) (*ast.Schema, error) {
	name, err := componentName(ref)
	if err != nil {
		return nil, err
	}

	return g.generateSchema(name, ogen.Schema{
		Ref: ref,
	})
}
