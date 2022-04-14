package gen

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/openapi"
)

func (g *Generator) generateRequest(ctx *genctx, opName string, body *openapi.RequestBody) (*ir.Request, error) {
	name := opName + "Req"

	contents, err := g.generateContents(ctx, name, !body.Required, body.Contents)
	if err != nil {
		return nil, errors.Wrap(err, "contents")
	}

	var requestType *ir.Type
	if len(contents) > 1 {
		requestType = ir.Interface(name)
		methodName, err := camel(name)
		if err != nil {
			return nil, errors.Wrap(err, "method name")
		}
		requestType.AddMethod(methodName)
		if err := ctx.saveType(requestType); err != nil {
			return nil, errors.Wrap(err, "save interface type")
		}
	}

	for contentType, t := range contents {
		switch {
		case len(contents) > 1:
			if !t.CanHaveMethods() {
				requestName, err := pascal(name, string(contentType))
				if err != nil {
					return nil, errors.Wrapf(err, "request name: %q", contentType)
				}
				t = ir.Alias(requestName, t)
				contents[contentType] = t
				if err := ctx.saveType(t); err != nil {
					return nil, errors.Wrap(err, "save alias type")
				}
			}

			t.Implement(requestType)

		case len(contents) == 1:
			requestType = t

		default:
			panic("unreachable")
		}
	}

	return &ir.Request{
		Type:     requestType,
		Contents: contents,
		Spec:     body,
	}, nil
}
