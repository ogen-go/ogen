package gen

import (
	"fmt"
	"strings"

	"github.com/ogen-go/ogen/internal/ast"
	"golang.org/x/xerrors"
)

func (g *Generator) resolveRequestBody(ref string) (*ast.RequestBody, error) {
	const prefix = "#/components/requestBodies/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, xerrors.Errorf("invalid requestBody reference: '%s'", ref)
	}

	componentName := strings.TrimPrefix(ref, prefix)
	name := pascal(componentName)
	if r, ok := g.requestBodies[name]; ok {
		return r, nil
	}

	component, found := g.spec.Components.RequestBodies[componentName]
	if !found {
		return nil, fmt.Errorf("component by reference '%s' not found", ref)
	}

	r, err := g.generateRequestBody(name, &component)
	if err != nil {
		return nil, err
	}

	g.requestBodies[name] = r
	return r, nil
}

func (g *Generator) resolveResponse(ref string) (*ast.Response, error) {
	const prefix = "#/components/responses/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, xerrors.Errorf("invalid response reference: '%s'", ref)
	}

	componentName := strings.TrimPrefix(ref, prefix)
	name := pascal(componentName)
	if r, ok := g.responses[name]; ok {
		// Example:
		//   ...
		//   responses:
		//     200:
		//       #/components/responses/Foo
		//     203:
		//       #/components/responses/Foo
		//
		// responses:
		//   Foo:
		//     contents:
		//       application/json:
		//         schema:
		//           type: string
		//
		// These responses (200, 203) in our ast representation
		// would point to the same *ast.Response struct.
		// It is bad because if we want to change response 200 *ast.Response.Contents map,
		// response 203 also changes, which can be unexpected.
		//
		// So, we need to create new *ast.Response and copy schemas into it.
		newR := ast.CreateResponse()
		newR.NoContent = r.NoContent
		newR.Contents = make(map[string]*ast.Schema)
		for ctype, s := range r.Contents {
			newR.Contents[ctype] = s
		}
		return newR, nil
	}

	component, found := g.spec.Components.Responses[componentName]
	if !found {
		return nil, fmt.Errorf("component by reference '%s' not found", ref)
	}

	r, err := g.generateResponse(name, component)
	if err != nil {
		return nil, err
	}

	g.responses[name] = r
	return r, nil
}

func (g *Generator) resolveParameter(ref string) (*ast.Parameter, error) {
	const prefix = "#/components/parameters/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, xerrors.Errorf("invalid parameter reference: '%s'", ref)
	}

	componentName := strings.TrimPrefix(ref, prefix)
	component, found := g.spec.Components.Parameters[componentName]
	if !found {
		return nil, fmt.Errorf("component by reference '%s' not found", ref)
	}

	return g.generateParameter(componentName, component)
}
