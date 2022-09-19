package gen

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/openapi"
)

func (g *Generator) generateRequest(ctx *genctx, opName string, body *openapi.RequestBody) (*ir.Request, error) {
	name := opName + "Req"
	optional := !body.Required

	contents, err := g.generateContents(ctx.appendPath("content"), name, optional, true, body.Content)
	if err != nil {
		return nil, errors.Wrap(err, "contents")
	}

	var (
		requestType *ir.Type
		isSumType   = len(contents) > 1 || optional
	)
	if isSumType {
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

	for contentType, content := range contents {
		t := content.Type
		switch {
		case isSumType:
			if !t.CanHaveMethods() {
				requestName, err := pascal(name, string(contentType))
				if err != nil {
					return nil, errors.Wrapf(err, "request name: %q", contentType)
				}
				t = ir.Alias(requestName, t)
				contents[contentType] = ir.Media{
					Encoding: content.Encoding,
					Type:     t,
				}
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

	var emptyBody *ir.Type
	if optional {
		if err := func() error {
			requestName, err := pascal(name, "EmptyBody")
			if err != nil {
				return errors.Wrapf(err, "generate name", requestName)
			}

			emptyBody = &ir.Type{
				Name: requestName,
				Kind: ir.KindStruct,
			}
			if err := ctx.saveType(emptyBody); err != nil {
				return errors.Wrap(err, "save type")
			}
			emptyBody.Implement(requestType)

			return nil
		}(); err != nil {
			return nil, errors.Wrap(err, "empty body type")
		}
	}

	return &ir.Request{
		Type:      requestType,
		EmptyBody: emptyBody,
		Contents:  contents,
		Spec:      body,
	}, nil
}
