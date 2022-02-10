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

func (g *Generator) generateResponses(opName string, responses map[string]*oas.Response) (*ir.Response, error) {
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
		r, err := g.responseToIR(respName, doc, resp)
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
		resp, err := g.responseToIR(respName, doc, def)
		if err != nil {
			return nil, errors.Wrap(err, "default")
		}

		result.Default = g.wrapResponseStatusCode(resp)
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
	g.saveIface(iface)
	walkResponseTypes(result, func(resName string, t *ir.Type) *ir.Type {
		switch t.Kind {
		case ir.KindPrimitive, ir.KindArray:
			t = ir.Alias(pascal(opName, resName), t)
			g.saveType(t)
		default:
		}

		t.Implement(iface)
		return t
	})

	result.Type = iface
	return result, nil
}

func (g *Generator) responseToIR(name, doc string, resp *oas.Response) (ret *ir.StatusResponse, err error) {
	if ref := resp.Ref; ref != "" {
		if r, ok := g.refs.responses[ref]; ok {
			return r, nil
		}

		name = pascal(strings.TrimPrefix(ref, "#/components/responses/"))
		doc = fmt.Sprintf("Ref: %s", ref)
		defer func() {
			if err == nil {
				g.refs.responses[ref] = ret
			}
		}()
	}

	if len(resp.Contents) == 0 {
		t := &ir.Type{
			Kind: ir.KindStruct,
			Name: name,
			Doc:  doc,
		}

		g.saveType(t)
		return &ir.StatusResponse{
			NoContent: t,
			Spec:      resp,
		}, nil
	}

	contents, err := g.generateContents(name, resp.Contents)
	if err != nil {
		return nil, errors.Wrap(err, "contents")
	}

	return &ir.StatusResponse{
		Contents: contents,
		Spec:     resp,
	}, nil
}

func (g *Generator) wrapResponseStatusCode(resp *ir.StatusResponse) (ret *ir.StatusResponse) {
	if ref := resp.Spec.Ref; ref != "" {
		if r, ok := g.wrapped.responses[ref]; ok {
			return r
		}
		defer func() { g.wrapped.responses[ref] = ret }()
	}

	if noc := resp.NoContent; noc != nil {
		if !noc.Is(ir.KindStruct, ir.KindAlias) {
			panic("unreachable")
		}

		return &ir.StatusResponse{
			Wrapped:   true,
			NoContent: g.wrapStatusCode(noc),
			Spec:      resp.Spec,
		}
	}

	contents := make(map[ir.ContentType]*ir.Type, len(resp.Contents))
	for contentType, t := range resp.Contents {
		contents[contentType] = g.wrapStatusCode(t)
	}

	return &ir.StatusResponse{
		Wrapped:  true,
		Contents: contents,
		Spec:     resp.Spec,
	}
}

func (g *Generator) wrapStatusCode(t *ir.Type) (ret *ir.Type) {
	if schema := t.Schema; schema != nil && schema.Ref != "" {
		if t, ok := g.wrapped.types[schema.Ref]; ok {
			return t
		}
		defer func() { g.wrapped.types[schema.Ref] = ret }()
	} else {
		defer func() { g.saveType(ret) }()
	}

	name := t.Name + "StatusCode"
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
	}
}
