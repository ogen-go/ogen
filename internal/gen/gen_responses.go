package gen

import (
	"fmt"
	"net/http"

	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen/internal/ir"
	"github.com/ogen-go/ogen/internal/oas"
)

func (g *Generator) generateResponses(opName string, responses *oas.OperationResponse) (*ir.Response, error) {
	name := opName + "Res"
	result := &ir.Response{
		Spec:       responses,
		StatusCode: map[int]*ir.StatusResponse{},
	}

	for code, resp := range responses.StatusCode {
		var (
			respName = pascal(name, http.StatusText(code))
			doc      = fmt.Sprintf("%s is response for %s operation.", respName, opName)
		)
		r, err := g.responseToIR(respName, doc, resp)
		if err != nil {
			return nil, xerrors.Errorf("%d: %w", code, err)
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
			return nil, xerrors.Errorf("default: %w", err)
		}

		for contentType, typ := range resp.Contents {
			resp.Contents[contentType] = g.wrapStatusCode(typ)
		}

		if typ := resp.NoContent; typ != nil {
			resp.NoContent = g.wrapStatusCode(typ)
		}

		result.Default = resp
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
		if typ.Is(ir.KindPrimitive, ir.KindArray) {
			typ = ir.Alias(pascal(opName, resName), typ)
			g.saveType(typ)
		}

		typ.Implement(iface)
		return typ
	})

	result.Type = iface
	return result, nil
}

func (g *Generator) responseToIR(name, doc string, resp *oas.Response) (*ir.StatusResponse, error) {
	if len(resp.Contents) == 0 {
		typ := &ir.Type{
			Kind: ir.KindStruct,
			Name: name,
			Doc:  doc,
		}

		g.saveType(typ)
		return &ir.StatusResponse{
			NoContent: typ,
			Spec:      resp,
		}, nil
	}

	types := make(map[ir.ContentType]*ir.Type)
	for contentType, schema := range resp.Contents {
		typ, err := g.generateSchema(pascal(name, contentType), schema)
		if err != nil {
			return nil, xerrors.Errorf("contents: %s: %w", contentType, err)
		}

		types[ir.ContentType(contentType)] = typ
	}
	return &ir.StatusResponse{
		Contents: types,
		Spec:     resp,
	}, nil
}

func (g *Generator) wrapStatusCode(typ *ir.Type) *ir.Type {
	if !typ.Is(ir.KindStruct, ir.KindAlias) {
		panic("unreachable")
	}

	name := typ.Name + "StatusCode"
	t := &ir.Type{
		Kind: ir.KindStruct,
		Name: name,
		Doc:  fmt.Sprintf("%s wraps %s with StatusCode.", name, typ.Name),
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
	}

	g.saveType(t)
	return t
}
