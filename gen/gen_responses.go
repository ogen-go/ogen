package gen

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/internal/oas"
)

func (g *Generator) generateResponses(ctx *genctx, opName string, responses map[string]*oas.Response) (*ir.Response, error) {
	name := opName + "Res"
	result := &ir.Response{
		StatusCode: make(map[int]*ir.StatusResponse, len(responses)),
	}

	// Sort responses by status code.
	statusCodes := make([]int, 0, len(responses))
	for status := range responses {
		switch status {
		case "default": // Ignore default response.
		default:
			// TODO: Support patterns like 5XX?
			code, err := strconv.Atoi(status)
			if err != nil {
				return nil, errors.Wrap(err, "parse response status code")
			}

			statusCodes = append(statusCodes, code)
		}
	}
	sort.Ints(statusCodes)

	for _, code := range statusCodes {
		var (
			resp     = responses[strconv.Itoa(code)]
			respName = pascal(opName, http.StatusText(code))
			doc      = fmt.Sprintf("%s is response for %s operation.", respName, opName)
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

		result.Default, err = g.wrapResponseStatusCode(ctx, respName, resp)
		if err != nil {
			return nil, errors.Wrap(err, "default")
		}
	}

	var (
		countTypes = 0
		lastWalked *ir.Type
	)

	walkResponseTypes(result, func(_ string, t *ir.Type) *ir.Type {
		countTypes += 1
		lastWalked = t
		return t
	})

	if countTypes == 1 {
		result.Type = lastWalked
		return result, nil
	}

	iface := ir.Interface(name)
	iface.AddMethod(camel(name))
	if err := ctx.saveType(iface); err != nil {
		return nil, errors.Wrap(err, "save interface type")
	}
	walkResponseTypes(result, func(resName string, t *ir.Type) *ir.Type {
		if !t.CanHaveMethods() {
			t = ir.Alias(pascal(opName, resName), t)
			if err := ctx.saveType(t); err != nil {
				panic("unreachable")
			}
		}

		t.Implement(iface)
		return t
	})

	result.Type = iface
	return result, nil
}

func (g *Generator) responseToIR(ctx *genctx, name, doc string, resp *oas.Response) (ret *ir.StatusResponse, err error) {
	if ref := resp.Ref; ref != "" {
		if r, ok := ctx.lookupResponse(ref); ok {
			return r, nil
		}

		name = pascal(strings.TrimPrefix(ref, "#/components/responses/"))
		doc = fmt.Sprintf("Ref: %s", ref)
		defer func() {
			if err == nil {
				ctx.local.responses[ref] = ret
			}
		}()
	}

	if len(resp.Contents) == 0 {
		t := &ir.Type{
			Kind: ir.KindStruct,
			Name: name,
			Doc:  doc,
		}

		if err := ctx.saveType(t); err != nil {
			return nil, errors.Wrap(err, "save type")
		}
		return &ir.StatusResponse{
			NoContent: t,
			Spec:      resp,
		}, nil
	}

	contents, err := g.generateContents(ctx, name, false, resp.Contents)
	if err != nil {
		return nil, errors.Wrap(err, "contents")
	}

	return &ir.StatusResponse{
		Contents: contents,
		Spec:     resp,
	}, nil
}

func (g *Generator) wrapResponseStatusCode(ctx *genctx, name string, resp *ir.StatusResponse) (ret *ir.StatusResponse, rerr error) {
	if ref := resp.Spec.Ref; ref != "" {
		if r, ok := ctx.lookupWrappedResponse(ref); ok {
			return r, nil
		}
		defer func() {
			if rerr != nil {
				return
			}

			ctx.local.wresponses[ref] = ret
		}()
	}

	if noc := resp.NoContent; noc != nil {
		w, err := g.wrapStatusCode(ctx, name, noc)
		if err != nil {
			return nil, err
		}

		return &ir.StatusResponse{
			Wrapped:   true,
			NoContent: w,
			Spec:      resp.Spec,
		}, nil
	}

	contents := make(map[ir.ContentType]*ir.Type, len(resp.Contents))
	for contentType, t := range resp.Contents {
		w, err := g.wrapStatusCode(ctx, name, t)
		if err != nil {
			return nil, err
		}

		contents[contentType] = w
	}

	return &ir.StatusResponse{
		Wrapped:  true,
		Contents: contents,
		Spec:     resp.Spec,
	}, nil
}

func (g *Generator) wrapStatusCode(ctx *genctx, name string, t *ir.Type) (ret *ir.Type, rerr error) {
	if schema := t.Schema; schema != nil && schema.Ref != "" {
		if t, ok := ctx.lookupWrappedType(schema.Ref); ok {
			return t, nil
		}

		defer func() { ctx.local.wtypes[schema.Ref] = ret }()
	} else {
		defer func() {
			if rerr != nil {
				return
			}

			rerr = ctx.saveType(ret)
		}()
	}

	if t.Name != "" {
		name = t.Name
	}

	name = name + "StatusCode"
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
