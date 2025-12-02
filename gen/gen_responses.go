package gen

import (
	"fmt"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/internal/xmaps"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
)

func (g *Generator) generateResponses(ctx *genctx, opName string, responses openapi.Responses) (_ *ir.Responses, err error) {
	name := opName + "Res"
	result := &ir.Responses{
		StatusCode: make(map[int]*ir.Response, len(responses.StatusCode)),
	}

	// Sort responses by status code.
	statusCodes := xmaps.SortedKeys(responses.StatusCode)
	for _, code := range statusCodes {
		respName, err := pascal(opName, statusText(code))
		if err != nil {
			return nil, errors.Wrapf(err, "%s: %d: response name", opName, code)
		}

		var (
			resp = responses.StatusCode[code]
			doc  = fmt.Sprintf("%s is response for %s operation.", respName, opName)
		)

		result.StatusCode[code], err = g.responseToIR(ctx, respName, doc, resp, false)
		if err != nil {
			return nil, errors.Wrapf(err, "code %d", code)
		}
	}

	for idx, resp := range responses.Pattern {
		if resp == nil {
			continue
		}
		pattern := fmt.Sprintf("%dXX", idx+1)

		respName, err := pascal(opName, pattern)
		if err != nil {
			return nil, errors.Wrapf(err, "%s: %s: response name", opName, pattern)
		}

		doc := fmt.Sprintf("%s is %s pattern response for %s operation.", respName, pattern, opName)

		result.Pattern[idx], err = g.responseToIR(ctx, respName, doc, resp, true)
		if err != nil {
			return nil, errors.Wrapf(err, "pattern %q", pattern)
		}
	}

	if def := responses.Default; def != nil && g.errType == nil {
		var (
			respName = opName + "Def"
			doc      = fmt.Sprintf("%s is default response for %s operation.", respName, opName)
		)

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

	// We'll need to use an interface-based approach for raw responses
	var needsInterface bool
	if countTypes == 1 {
	statusCodesLoop:
		for _, resp := range result.StatusCode {
			for _, media := range resp.Contents {
				if media.RawResponse {
					needsInterface = true
					break statusCodesLoop
				}
			}
		}
		if !needsInterface {
		patternLoop:
			for _, resp := range result.Pattern {
				if resp == nil {
					continue
				}
				for _, media := range resp.Contents {
					if media.RawResponse {
						needsInterface = true
						break patternLoop
					}
				}
			}
		}
		if !needsInterface && result.Default != nil {
			for _, media := range result.Default.Contents {
				if media.RawResponse {
					needsInterface = true
					break
				}
			}
		}
	}

	if countTypes == 1 && !needsInterface {
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

	// Add raw response concrete types for content types with RawResponse=true
	if err := addRawResponseTypes(ctx, result, iface, opName); err != nil {
		return nil, errors.Wrap(err, "add raw response types")
	}

	result.Type = iface
	return result, nil
}

// addRawResponseTypes adds concrete types for raw responses that implement the interface
func addRawResponseTypes(ctx *genctx, result *ir.Responses, iface *ir.Type, opName string) error {
	addRawType := func(prefix string, response *ir.Response) error {
		if response == nil {
			return nil
		}

		for contentType, media := range response.Contents {
			if !media.RawResponse {
				continue
			}

			rawTypeName, err := pascal(opName, prefix, "Raw", string(contentType))
			if err != nil {
				return errors.Wrapf(err, "raw type name: %s %s", prefix, contentType)
			}

			rawType := &ir.Type{
				Kind: ir.KindStruct,
				Name: rawTypeName,
				Doc:  fmt.Sprintf("%s represents raw HTTP response for %s %s.", rawTypeName, opName, contentType),
				Fields: []*ir.Field{
					{
						Name: "Response",
						Type: ir.Pointer(&ir.Type{
							Kind:      ir.KindPrimitive,
							Primitive: "http.Response",
							External: ir.ExternalType{
								PackagePath: "net/http",
								TypeName:    "Response",
								IsPointer:   false,
							},
						}, ir.NilOptional),
						Tag: ir.Tag{JSON: "-"},
					},
				},
			}

			// Remove the original structured type from the interface
			// since we're replacing it with a raw response type
			originalType := media.Type
			originalType.Unimplement(iface)

			rawType.Implement(iface)

			if err := ctx.saveType(rawType); err != nil {
				return errors.Wrap(err, "save raw type")
			}

			response.Contents[contentType] = ir.Media{
				Encoding:      media.Encoding,
				Type:          rawType,
				JSONStreaming: media.JSONStreaming,
				RawResponse:   media.RawResponse,
			}
		}
		return nil
	}

	for code, response := range result.StatusCode {
		if err := addRawType(statusText(code), response); err != nil {
			return errors.Wrapf(err, "status code %d", code)
		}
	}

	for pattern, response := range result.Pattern {
		if err := addRawType(fmt.Sprintf("%dXX", pattern+1), response); err != nil {
			return errors.Wrapf(err, "pattern %d", pattern)
		}
	}

	if err := addRawType("Default", result.Default); err != nil {
		return errors.Wrap(err, "default")
	}

	return nil
}

func (g *Generator) responseToIR(
	ctx *genctx,
	name, doc string,
	resp *openapi.Response,
	withStatusCode bool,
) (ret *ir.Response, rerr error) {
	if ref := resp.Ref; !ref.IsZero() {
		if r, ok := ctx.lookupResponse(ref); ok {
			return r, nil
		}

		n, err := pascal(cleanRef(ref))
		if err != nil {
			return nil, errors.Wrapf(err, "response name: %q", ref)
		}
		name = n
		doc = fmt.Sprintf("Ref: %s", ref.Ptr)
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

	headers, err := g.generateHeaders(ctx, name, resp.Headers)
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

	contents, err := g.generateContents(ctx, name, false, false, resp.Content)
	if err != nil {
		return nil, errors.Wrap(err, "contents")
	}

	// Check for unsupported response content types.
	var unsupported []string
	for ct, content := range contents {
		t, e := content.Type, content.Encoding
		if e.JSON() || e.ProblemJSON() || t.IsStream() || isBinary(t.Schema) || content.RawResponse {
			continue
		}
		delete(contents, ct)
		unsupported = append(unsupported, ct.String())
	}
	if len(contents) == 0 && len(unsupported) > 0 {
		return nil, &ErrUnsupportedContentTypes{ContentTypes: unsupported}
	}

	for contentType, media := range contents {
		if contentType.Mask() {
			if headers == nil {
				headers = map[string]*ir.Parameter{}
			}
			headers["Content-Type"] = &ir.Parameter{
				Name: "ContentType",
				Type: ir.Primitive(ir.String, nil),
				Spec: &openapi.Parameter{
					Name:     "Content-Type",
					In:       openapi.LocationHeader,
					Required: true,
				},
			}
		}
		// Use content-type-specific name for wrapper when there are multiple contents
		// to avoid name conflicts (e.g., when both application/json and
		// application/vnd.github.v3.star+json have array schemas without names).
		wrapperName := name
		if len(contents) > 1 {
			wrapperName, _ = pascal(name, string(contentType))
		}
		t, err := wrapResponseType(ctx, wrapperName, resp.Ref, media.Type, headers, withStatusCode, len(contents) > 1)
		if err != nil {
			return nil, errors.Wrapf(err, "content: %q: wrap response type", contentType)
		}
		contents[contentType] = ir.Media{
			Encoding:      media.Encoding,
			Type:          t,
			JSONStreaming: media.JSONStreaming,
			RawResponse:   media.RawResponse,
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
	respRef jsonschema.Ref,
	t *ir.Type,
	headers map[string]*ir.Parameter,
	withStatusCode bool,
	multipleContents bool,
) (ret *ir.Type, rerr error) {
	if len(headers) == 0 && !withStatusCode {
		return t, nil
	}

	if schema := t.Schema; schema != nil && !schema.Ref.IsZero() {
		if t, ok := ctx.lookupWType(respRef, schema.Ref); ok {
			return t, nil
		}

		defer func() {
			if rerr != nil {
				return
			}

			if err := ctx.saveWType(respRef, schema.Ref, ret); err != nil {
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

	// Prefer response name to schema name in case of wrapping.
	if (respRef.IsZero() || multipleContents) && t.Name != "" {
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

	for _, key := range xmaps.SortedKeys(headers) {
		h := headers[key]
		t.Fields = append(t.Fields, &ir.Field{
			Name: h.Name,
			Type: h.Type,
		})
	}
}
