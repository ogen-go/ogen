package gen

import (
	"golang.org/x/xerrors"

	ast "github.com/ogen-go/ogen/internal/ast"
	"github.com/ogen-go/ogen/internal/ir"
)

func (g *Generator) generateRequest(name string, body *ast.RequestBody) (*ir.Request, error) {
	types := make(map[string]*ir.Type)
	for contentType, schema := range body.Contents {
		sName := name
		if len(body.Contents) > 1 {
			sName = pascal(name, contentType)
		}

		typ, err := g.generateSchema(sName, schema)
		if err != nil {
			return nil, xerrors.Errorf("contents: %s: %w", contentType, err)
		}

		types[contentType] = typ
	}

	if len(types) == 1 {
		for _, typ := range types {
			return &ir.Request{
				Type:     typ,
				Contents: types,
				Required: body.Required,
				Spec:     body,
			}, nil
		}
	}

	iface := ir.Interface(name)
	iface.AddMethod(camel(name))
	g.saveIface(iface)
	for contentType, typ := range types {
		if typ.Is(ir.KindPrimitive, ir.KindArray) {
			// Primitive types cannot have methods, wrap it with alias.
			typ = ir.Alias(pascal(name, contentType), typ)
			g.saveType(typ)
		}

		typ.Implement(iface)
	}

	return &ir.Request{
		Type:     iface,
		Contents: types,
		Required: body.Required,
		Spec:     body,
	}, nil
}
