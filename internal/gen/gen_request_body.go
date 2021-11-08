package gen

import (
	"github.com/go-faster/errors"

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

		if isBinary(schema) {
			types[ir.ContentType(contentType)] = ir.Stream()
			continue
		}

		if schema == nil {
			switch contentType {
			case "application/octet-stream":
				typ := ir.Stream()
				types[ir.ContentType(contentType)] = typ
				g.saveType(typ)
				continue
			default:
				return nil, errors.Errorf("unsupported empty schema for content-type %q", contentType)
			}
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
		switch typ.Kind {
		case ir.KindPrimitive, ir.KindArray:
			// Primitive types cannot have methods, wrap it with alias.
			typ = ir.Alias(pascal(name, string(contentType)), typ)
			types[contentType] = typ
			g.saveType(typ)
		case ir.KindStream:
			typ.Name = pascal(name, string(contentType))
			g.saveType(typ)
		default:
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
