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
				t := ir.Stream()
				types[ir.ContentType(contentType)] = t
				g.saveType(t)
				continue
			default:
				return nil, errors.Errorf("unsupported empty schema for content-type %q", contentType)
			}
		}

		t, err := g.generateSchema(sName, schema)
		if err != nil {
			return nil, errors.Wrapf(err, "contents: %s", contentType)
		}
		if !body.Required {
			t = ir.Generic(genericPostfix(t.Go()), t, ir.GenericVariant{
				Optional: true,
			})
			g.saveType(t)
		}

		types[ir.ContentType(contentType)] = t
	}

	if len(types) == 1 {
		for _, t := range types {
			return &ir.Request{
				Type:     t,
				Contents: types,
				Spec:     body,
			}, nil
		}
	}

	iface := ir.Interface(name)
	iface.AddMethod(camel(name))
	g.saveIface(iface)
	for contentType, t := range types {
		switch t.Kind {
		case ir.KindPrimitive, ir.KindArray:
			// Primitive types cannot have methods, wrap it with alias.
			t = ir.Alias(pascal(name, string(contentType)), t)
			types[contentType] = t
			g.saveType(t)
		case ir.KindStream:
			t.Name = pascal(name, string(contentType))
			g.saveType(t)
		default:
		}

		t.Implement(iface)
	}

	return &ir.Request{
		Type:     iface,
		Contents: types,
		Spec:     body,
	}, nil
}
