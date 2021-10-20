package gen

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/ast"
)

func (g *Generator) generateMethods() error {
	for path, item := range g.spec.Paths {
		if item.Ref != "" {
			return xerrors.New("referenced paths are not supported")
		}

		if g.opt.SpecificMethodPath != "" {
			if g.opt.SpecificMethodPath != path {
				continue
			}
		}

		if err := forEachOps(item, func(method string, op ogen.Operation) error {
			if err := g.generateMethod(path, strings.ToUpper(method), op); err != nil {
				if err := g.checkErr(err); err != nil {
					return xerrors.Errorf("%s: %w", method, err)
				}
			}
			return nil
		}); err != nil {
			return xerrors.Errorf("paths: %s: %w", path, err)
		}
	}

	sort.SliceStable(g.methods, func(i, j int) bool {
		return strings.Compare(g.methods[i].RawPath, g.methods[j].RawPath) < 0
	})

	return nil
}

func (g *Generator) generateMethod(path, method string, op ogen.Operation) (err error) {
	// Use path + method as unique identifier.
	methodName := pascal(path, strings.ToLower(method))
	if op.OperationID != "" {
		// Use operationId if present.
		methodName = pascal(op.OperationID)
	}

	m := &ast.Method{
		OperationID: op.OperationID,
		Name:        methodName,
		RawPath:     path,
		HTTPMethod:  method,
	}

	m.Parameters, err = g.generateParams(m.Name, op.Parameters)
	if err != nil {
		return xerrors.Errorf("parameters: %w", err)
	}

	m.PathParts, err = parsePath(path, m.PathParams())
	if err != nil {
		return xerrors.Errorf("parse path: %w", err)
	}

	if op.RequestBody != nil {
		iface := ast.Iface(methodName + "Request")
		iface.AddMethod(camel(methodName + "Request"))
		iface.SetDoc(fmt.Sprintf("%s represents %s request.", iface.Name, op.OperationID))
		g.interfaces[iface.Name] = iface

		rbody, err := g.generateRequestBody(methodName, op.RequestBody)
		if err != nil {
			return xerrors.Errorf("requestBody: %w", err)
		}

		for _, schema := range rbody.Contents {
			schema.Implement(iface)
		}

		m.RequestBody = rbody
		m.RequestType = iface
	}

	if len(op.Responses) > 0 {
		iface := ast.Iface(methodName + "Response")
		iface.SetDoc(fmt.Sprintf("%s represents %s response.", iface.Name, op.OperationID))
		iface.AddMethod(camel(methodName + "Response"))
		g.interfaces[iface.Name] = iface

		responses, err := g.generateResponses(methodName, op.Responses)
		if err != nil {
			return xerrors.Errorf("responses: %w", err)
		}

		for _, resp := range responses.StatusCode {
			resp.Implement(iface)
		}

		if def := responses.Default; def != nil {
			def.Implement(iface)
		}

		m.Responses = responses
		m.ResponseType = iface
	}

	g.methods = append(g.methods, m)
	return nil
}
