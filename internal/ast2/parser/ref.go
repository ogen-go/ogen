package parser

import (
	"fmt"
	"strings"

	ast "github.com/ogen-go/ogen/internal/ast2"
	"golang.org/x/xerrors"
)

func (p *parser) resolveRequestBody(ref string) (*ast.RequestBody, error) {
	const prefix = "#/components/requestBodies/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, xerrors.Errorf("invalid requestBody reference: '%s'", ref)
	}

	if r, ok := p.refs.requestBodies[ref]; ok {
		return r, nil
	}

	componentName := strings.TrimPrefix(ref, prefix)
	component, found := p.spec.Components.RequestBodies[componentName]
	if !found {
		return nil, fmt.Errorf("component by reference '%s' not found", ref)
	}

	r, err := p.parseRequestBody(&component)
	if err != nil {
		return nil, err
	}

	p.refs.requestBodies[ref] = r
	return r, nil
}

func (p *parser) resolveResponse(ref string) (*ast.Response, error) {
	const prefix = "#/components/responses/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, xerrors.Errorf("invalid response reference: '%s'", ref)
	}

	if r, ok := p.refs.responses[ref]; ok {
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

	componentName := strings.TrimPrefix(ref, prefix)
	component, found := p.spec.Components.Responses[componentName]
	if !found {
		return nil, fmt.Errorf("component by reference '%s' not found", ref)
	}

	r, err := p.parseResponse(component)
	if err != nil {
		return nil, err
	}

	p.refs.responses[ref] = r
	return r, nil
}

func (p *parser) resolveParameter(ref string) (*ast.Parameter, error) {
	const prefix = "#/components/parameters/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, xerrors.Errorf("invalid parameter reference: '%s'", ref)
	}

	componentName := strings.TrimPrefix(ref, prefix)
	component, found := p.spec.Components.Parameters[componentName]
	if !found {
		return nil, fmt.Errorf("component by reference '%s' not found", ref)
	}

	return p.parseParameter(component)
}
