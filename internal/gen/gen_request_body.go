package gen

import (
	"github.com/ogen-go/errors"

	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/internal/oas"
)

func (g *Generator) generateRequest(opName string, body *oas.RequestBody) (*ir.Request, error) {
	var (
		name  = opName + "Req"
		types = make(map[ir.ContentType]*ir.Type)
	)

	contentTypes, err := sortContentTypes(body.Contents)
	if err != nil {
		return nil, err
	}

	for _, contentType := range contentTypes {
		var (
			schema = body.Contents[contentType]
			sName  = name
		)
		if len(body.Contents) > 1 {
			sName = pascal(name, contentType)
		}

		typ, err := g.generateSchema(sName, schema)
		if err != nil {
			return nil, errors.Wrapf(err, "contents: %s", contentType)
		}

		types[ir.ContentType(contentType)] = typ
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
			typ = ir.Alias(pascal(name, string(contentType)), typ)
			types[contentType] = typ
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
