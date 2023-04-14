package gen

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/openapi"
)

func (g *Generator) generateRequest(ctx *genctx, opName string, body *openapi.RequestBody) (*ir.Request, error) {
	name := opName + "Req"

	// Filter early to check the number of media below.
	//
	// FIXME(tdakkota): do not modify the original body.
	rawContents := body.Content
	if err := filterMostSpecific(rawContents, g.log); err != nil {
		return nil, errors.Wrap(err, "filter most specific")
	}

	// Generate optional type only if there is only one media type and body is not required.
	//
	// Otherwise, we generate a special "EmptyBody" case.
	generateOptional := len(rawContents) == 1 && !body.Required
	contents, err := g.generateContents(ctx, name, generateOptional, true, rawContents)
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

	for contentType, content := range contents {
		t := content.Type
		switch {
		case len(contents) > 1:
			if !t.CanHaveMethods() {
				requestName, err := pascal(name, contentType.String())
				if err != nil {
					return nil, errors.Wrapf(err, "request name: %q", contentType)
				}
				t = ir.Alias(requestName, t)
				contents[contentType] = ir.Media{
					Encoding:      content.Encoding,
					Type:          t,
					JSONStreaming: content.JSONStreaming,
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

	// Generate an empty body case only if there is more than one media type.
	//
	// If there is only one media type, we generate an optional type instead.
	var emptyBody *ir.Type
	if len(contents) > 1 && !body.Required {
		if err := func() error {
			requestName, err := pascal(name, "EmptyBody")
			if err != nil {
				return errors.Wrap(err, "generate name")
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
