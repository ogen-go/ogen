package gen

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/internal/oas"
)

func (g *Generator) generateResponses(opName string, responses *oas.OperationResponse) (*ir.Response, error) {
	name := opName + "Res"
	result := &ir.Response{
		Spec:       responses,
		StatusCode: map[int]*ir.StatusResponse{},
	}

	statusCodes := make([]int, 0, len(responses.StatusCode))
	for code := range responses.StatusCode {
		statusCodes = append(statusCodes, code)
	}
	sort.Ints(statusCodes)

	for _, code := range statusCodes {
		var (
			resp     = responses.StatusCode[code]
			respName = pascal(opName, http.StatusText(code))
			doc      = fmt.Sprintf("%s is response for %s operation.", respName, opName)
		)
		r, err := g.responseToIR(respName, doc, resp)
		if err != nil {
			return nil, errors.Wrapf(err, "%d", code)
		}

		result.StatusCode[code] = r
	}

	if def := responses.Default; def != nil {
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

	walkResponseTypes(result, func(name string, typ *ir.Type) *ir.Type {
		countTypes += 1
		lastWalked = typ
		return typ
	})

	if countTypes == 1 {
		result.Type = lastWalked
		return result, nil
	}

	iface := ir.Interface(name)
	iface.AddMethod(camel(name))
	g.saveIface(iface)
	walkResponseTypes(result, func(resName string, typ *ir.Type) *ir.Type {
		switch typ.Kind {
		case ir.KindPrimitive, ir.KindArray:
			typ = ir.Alias(pascal(opName, resName), typ)
			g.saveType(typ)
		case ir.KindStream:
			typ.Stream.Name = pascal(opName, resName)
			g.saveType(typ)
		}

		typ.Implement(iface)
		return typ
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
		typ := &ir.Type{
			Doc:  doc,
			Kind: ir.KindStruct,
			Struct: &ir.TypeStruct{
				Name: name,
			},
		}

		g.saveType(typ)
		return &ir.StatusResponse{
			NoContent: typ,
			Spec:      resp,
		}, nil
	}

	types := make(map[ir.ContentType]*ir.Type, len(resp.Contents))

	contentTypes, err := sortContentTypes(resp.Contents)
	if err != nil {
		return nil, err
	}

	for _, contentType := range contentTypes {
		schema := resp.Contents[contentType]
		if isBinary(schema) {
			types[ir.ContentType(contentType)] = ir.Stream()
			continue
		}

		if schema == nil {
			switch contentType {
			case "application/octet-stream":
				types[ir.ContentType(contentType)] = ir.Stream()
				continue
			default:
				return nil, errors.Errorf("unsupported empty schema for content-type %q", contentType)
			}
		}

		typeName := name
		if len(resp.Contents) > 1 {
			typeName = pascal(name, contentType)
		}

		typ, err := g.generateSchema(typeName, schema)
		if err != nil {
			return nil, errors.Wrapf(err, "contents: %s", contentType)
		}

		types[ir.ContentType(contentType)] = typ
	}
	return &ir.StatusResponse{
		Contents: types,
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
	for contentType, typ := range resp.Contents {
		contents[contentType] = g.wrapStatusCode(typ)
	}

	return &ir.StatusResponse{
		Wrapped:  true,
		Contents: contents,
		Spec:     resp.Spec,
	}
}

func (g *Generator) wrapStatusCode(typ *ir.Type) (ret *ir.Type) {
	typeName := ""
	switch typ.Kind {
	case ir.KindStruct:
		typeName = typ.Struct.Name
	case ir.KindAlias:
		typeName = typ.Alias.Name
	default:
		panic("unreachable")
	}

	if schema := typ.Schema(); schema != nil && schema.Ref != "" {
		if t, ok := g.wrapped.types[schema.Ref]; ok {
			return t
		}
		defer func() { g.wrapped.types[schema.Ref] = ret }()
	} else {
		defer func() { g.saveType(ret) }()
	}

	name := typeName + "StatusCode"
	return &ir.Type{
		Doc:  fmt.Sprintf("%s wraps %s with StatusCode.", name, typeName),
		Kind: ir.KindStruct,
		Struct: &ir.TypeStruct{
			Name: name,
			Fields: []*ir.Field{
				{
					Name: "StatusCode",
					Type: ir.Primitive(ir.Int, nil),
				},
				{
					Name: "Response",
					Type: typ,
				},
			},
		},
	}
}
