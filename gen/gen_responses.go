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

func (g *Generator) generateResponses(ctx *genctx, opName string, responses map[string]*openapi.Response) (_ *ir.Responses, err error) {
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
		ctx := ctx.appendPath(strconv.Itoa(code))
		result.StatusCode[code], err = g.responseToIR(ctx, respName, doc, resp, false)
		if err != nil {
			return nil, errors.Wrapf(err, "%d", code)
		}
	}

	if def, ok := responses["default"]; ok && g.errType == nil {
		var (
			respName = opName + "Def"
			doc      = fmt.Sprintf("%s is default response for %s operation.", respName, opName)
		)
		ctx := ctx.appendPath("default")
		result.Default, err = g.responseToIR(ctx, respName, doc, def, true)
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

func (g *Generator) responseToIR(
	ctx *genctx,
	name, doc string,
	resp *openapi.Response,
	withStatusCode bool,
) (ret *ir.Response, rerr error) {
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

	headers, err := g.generateHeaders(ctx.appendPath("headers"), name, resp.Headers)
	if err != nil {
		return nil, errors.Wrap(err, "headers")
	}

	if len(resp.Content) == 0 {
		t := &ir.Type{
			Kind: ir.KindStruct,
			Name: name,
			Doc:  doc,
		}

		injectHeaderFields(headers, t)
		if withStatusCode {
			t.Fields = append(t.Fields, &ir.Field{
				Name: "StatusCode",
				Type: ir.Primitive(ir.Int, nil),
			})
		}

		if err := ctx.saveType(t); err != nil {
			return nil, errors.Wrap(err, "save type")
		}

		return &ir.Response{
			NoContent:      t,
			Spec:           resp,
			Headers:        headers,
			WithStatusCode: withStatusCode,
			WithHeaders:    len(headers) > 0,
		}, nil
	}

	contents, err := g.generateContents(ctx.appendPath("content"), name, false, resp.Content)
	if err != nil {
		return nil, errors.Wrap(err, "contents")
	}

	// Check for unsupported response content types.
	var unsupported []string
	for ct, content := range contents {
		t, e := content.Type, content.Encoding
		if e.JSON() || t.IsStream() || isBinary(t.Schema) {
			continue
		}
		delete(contents, ct)
		unsupported = append(unsupported, string(ct))
	}
	if len(contents) == 0 && len(unsupported) > 0 {
		return nil, &ErrUnsupportedContentTypes{ContentTypes: unsupported}
	}

	for contentType, media := range contents {
		t, err := wrapResponseType(ctx, name, media.Type, headers, withStatusCode)
		if err != nil {
			return nil, errors.Wrapf(err, "content: %q: wrap response type", contentType)
		}
		contents[contentType] = ir.Media{
			Encoding: media.Encoding,
			Type:     t,
		}
	}

	return &ir.Response{
		Contents:       contents,
		Spec:           resp,
		Headers:        headers,
		WithStatusCode: withStatusCode,
		WithHeaders:    len(headers) > 0,
	}, nil
}

func wrapResponseType(
	ctx *genctx,
	name string,
	t *ir.Type,
	headers map[string]*ir.Parameter,
	withStatusCode bool,
) (ret *ir.Type, rerr error) {
	if len(headers) == 0 && !withStatusCode {
		return t, nil
	}

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

	var (
		namePostfix string
		doc         string
	)
	switch {
	case len(headers) > 0 && withStatusCode:
		namePostfix = "StatusCodeWithHeaders"
		doc = fmt.Sprintf("%sStatusCodeWithHeaders wraps %s with status code and response headers.", name, t.Go())
	case len(headers) > 0:
		namePostfix = "Headers"
		doc = fmt.Sprintf("%sHeaders wraps %s with response headers.", name, t.Go())
	case withStatusCode:
		namePostfix = "StatusCode"
		doc = fmt.Sprintf("%sStatusCode wraps %s with StatusCode.", name, t.Go())
	default:
		panic("unreachable")
	}

	wrapper := &ir.Type{
		Kind: ir.KindStruct,
		Name: name + namePostfix,
		Doc:  doc,
	}

	if withStatusCode {
		wrapper.Fields = append(wrapper.Fields, &ir.Field{
			Name: "StatusCode",
			Type: ir.Primitive(ir.Int, nil),
		})
	}

	injectHeaderFields(headers, wrapper)
	wrapper.Fields = append(wrapper.Fields, &ir.Field{
		Name: "Response",
		Type: t,
	})

	return wrapper, nil
}

func injectHeaderFields(headers map[string]*ir.Parameter, t *ir.Type) {
	if !t.IsStruct() {
		panic(fmt.Sprintf("expected struct, got %q", t.Kind))
	}

	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, key := range keys {
		h := headers[key]
		t.Fields = append(t.Fields, &ir.Field{
			Name: h.Name,
			Type: h.Type,
		})
	}
}
