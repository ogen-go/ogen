package gen

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/openapi"
)

func (g *Generator) generateResponses(ctx *genctx, opName string, responses map[string]*openapi.Response) (*ir.Responses, error) {
	name := opName + "Res"
	result := &ir.Responses{
		StatusCode: make(map[int]*ir.Response, len(responses)),
	}

	// Sort responses by status code.
	statusCodes := make([]int, 0, len(responses))
	for status := range responses {
		if status == "default" {
			continue // Ignore default response.
		}
		switch strings.ToUpper(status) {
		case "1XX", "2XX", "3XX", "4XX", "5XX":
			return nil, &ErrNotImplemented{Name: "HTTP code pattern"}
		default:
			code, err := strconv.Atoi(status)
			if err != nil {
				return nil, errors.Wrap(err, "parse response status code")
			}

			statusCodes = append(statusCodes, code)
		}
	}
	sort.Ints(statusCodes)

	for _, code := range statusCodes {
		respName, err := pascal(opName, statusText(code))
		if err != nil {
			return nil, errors.Wrapf(err, "%s: %d: response name", opName, code)
		}

		var (
			resp = responses[strconv.Itoa(code)]
			doc  = fmt.Sprintf("%s is response for %s operation.", respName, opName)
		)
		r, err := g.responseToIR(ctx, respName, doc, resp)
		if err != nil {
			return nil, errors.Wrapf(err, "%d", code)
		}

		result.StatusCode[code] = r
	}

	if def, ok := responses["default"]; ok && g.errType == nil {
		var (
			respName = opName + "Def"
			doc      = fmt.Sprintf("%s is default response for %s operation.", respName, opName)
		)
		resp, err := g.responseToIR(ctx, respName, doc, def)
		if err != nil {
			return nil, errors.Wrap(err, "default")
		}

		result.Default, err = wrapResponseStatusCode(ctx, respName, resp)
		if err != nil {
			return nil, errors.Wrap(err, "default")
		}
	}

	var (
		countTypes = 0
		lastWalked *ir.Type
	)

	if err := walkResponseTypes(result, func(_ string, t *ir.Type) (*ir.Type, error) {
		countTypes++
		lastWalked = t
		return t, nil
	}); err != nil {
		return nil, errors.Wrap(err, "walk")
	}

	if countTypes == 1 {
		result.Type = lastWalked
		return result, nil
	}

	iface := ir.Interface(name)
	methodName, err := camel(name)
	if err != nil {
		return nil, errors.Wrap(err, "method name")
	}
	iface.AddMethod(methodName)
	if err := ctx.saveType(iface); err != nil {
		return nil, errors.Wrap(err, "save interface type")
	}
	if err := walkResponseTypes(result, func(resName string, t *ir.Type) (*ir.Type, error) {
		if !t.CanHaveMethods() {
			respName, err := pascal(opName, resName)
			if err != nil {
				return nil, errors.Wrapf(err, "request name: %q", resName)
			}
			t = ir.Alias(respName, t)
			if err := ctx.saveType(t); err != nil {
				return nil, errors.Wrap(err, "save type")
			}
		}

		t.Implement(iface)
		return t, nil
	}); err != nil {
		return nil, errors.Wrap(err, "walk")
	}

	result.Type = iface
	return result, nil
}

func (g *Generator) responseToIR(ctx *genctx, name, doc string, resp *openapi.Response) (ret *ir.Response, rerr error) {
	if ref := resp.Ref; ref != "" {
		if r, ok := ctx.lookupResponse(ref); ok {
			return r, nil
		}

		n, err := pascal(strings.TrimPrefix(ref, "#/components/responses/"))
		if err != nil {
			return nil, errors.Wrapf(err, "response name: %q", ref)
		}
		name = n
		doc = fmt.Sprintf("Ref: %s", ref)
		defer func() {
			if rerr != nil {
				return
			}

			if err := ctx.saveResponse(ref, ret); err != nil {
				rerr = err
				ret = nil
			}
		}()
	}

	if len(resp.Content) == 0 {
		t := &ir.Type{
			Kind: ir.KindStruct,
			Name: name,
			Doc:  doc,
		}

		if err := ctx.saveType(t); err != nil {
			return nil, errors.Wrap(err, "save type")
		}
		return &ir.Response{
			NoContent: t,
			Spec:      resp,
		}, nil
	}

	contents, err := g.generateContents(ctx, name, false, resp.Content)
	if err != nil {
		return nil, errors.Wrap(err, "contents")
	}

	// Check for unsupported response content types.
	var unsupported []string
	for ct, content := range contents {
		if content.IsStream() || isBinary(content.Schema) {
			continue
		}
		switch ct {
		case ir.ContentTypeJSON, ir.ContentTypeOctetStream:
		default:
			delete(contents, ct)
			unsupported = append(unsupported, string(ct))
		}
	}
	if len(contents) == 0 && len(unsupported) > 0 {
		return nil, &ErrUnsupportedContentTypes{ContentTypes: unsupported}
	}

	return &ir.Response{
		Contents: contents,
		Spec:     resp,
	}, nil
}

func wrapResponseStatusCode(ctx *genctx, name string, resp *ir.Response) (ret *ir.Response, rerr error) {
	if ref := resp.Spec.Ref; ref != "" {
		if r, ok := ctx.lookupWrappedResponse(ref); ok {
			return r, nil
		}
		defer func() {
			if rerr != nil {
				return
			}

			if err := ctx.saveWResponse(ref, ret); err != nil {
				rerr = err
				ret = nil
			}
		}()
	}

	if noc := resp.NoContent; noc != nil {
		w, err := wrapStatusCode(ctx, name, noc)
		if err != nil {
			return nil, err
		}

		return &ir.Response{
			Wrapped:   true,
			NoContent: w,
			Spec:      resp.Spec,
		}, nil
	}

	contents := make(map[ir.ContentType]*ir.Type, len(resp.Contents))
	for contentType, t := range resp.Contents {
		w, err := wrapStatusCode(ctx, name, t)
		if err != nil {
			return nil, err
		}

		contents[contentType] = w
	}

	return &ir.Response{
		Wrapped:  true,
		Contents: contents,
		Spec:     resp.Spec,
	}, nil
}

func wrapStatusCode(ctx *genctx, name string, t *ir.Type) (ret *ir.Type, rerr error) {
	if schema := t.Schema; schema != nil && schema.Ref != "" {
		if t, ok := ctx.lookupWType(schema.Ref); ok {
			return t, nil
		}

		defer func() {
			if rerr != nil {
				return
			}

			if err := ctx.saveWType(schema.Ref, ret); err != nil {
				rerr = err
				ret = nil
			}
		}()
	} else {
		defer func() {
			if rerr != nil {
				return
			}

			if err := ctx.saveType(ret); err != nil {
				rerr = err
				ret = nil
			}
		}()
	}

	if t.Name != "" {
		name = t.Name
	}

	name += "StatusCode"
	return &ir.Type{
		Kind: ir.KindStruct,
		Name: name,
		Doc:  fmt.Sprintf("%s wraps %s with StatusCode.", name, t.Name),
		Fields: []*ir.Field{
			{
				Name: "StatusCode",
				Type: ir.Primitive(ir.Int, nil),
			},
			{
				Name: "Response",
				Type: t,
			},
		},
	}, nil
}
