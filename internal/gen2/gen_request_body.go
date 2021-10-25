package gen

import (
	ast "github.com/ogen-go/ogen/internal/ast2"
	"github.com/ogen-go/ogen/internal/ir"
	"golang.org/x/xerrors"
)

func (g *Generator) generateRequest(name string, body *ast.RequestBody) (*ir.Request, error) {
	types := make(map[string]*ir.Type)
	for contentType, schema := range body.Contents {
		sname := name
		if len(body.Contents) > 1 {
			sname = pascal(name, contentType)
		}

		typ, err := g.generateSchema(sname, schema)
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

	iface := ir.Iface(name)
	iface.AddMethod(camel(name))
	g.types[iface.Name] = iface
	for contentType, typ := range types {
		if typ.Is(ir.KindPrimitive, ir.KindArray) {
			// Primitive types cannot have methods, wrap it with alias.
			typ = ir.Alias(pascal(name, contentType), typ)
		}

		typ.Implement(iface)
		g.types[typ.Name] = typ
	}

	return &ir.Request{
		Type:     iface,
		Contents: types,
		Required: body.Required,
		Spec:     body,
	}, nil
}
