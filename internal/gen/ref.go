package gen

import (
	"fmt"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/ast"
)

func (g *Generator) resolveSchema(ref string) (*ast.Schema, error) {
	name, err := componentName(ref)
	if err != nil {
		return nil, err
	}

	return g.generateSchema(name, ogen.Schema{
		Ref: ref,
	})
}

func (g *Generator) resolveRequestBody(ref string) (*ast.RequestBody, error) {
	name, err := componentName(ref)
	if err != nil {
		return nil, err
	}

	if r, ok := g.requestBodies[name]; ok {
		return r, nil
	}

	component, found := g.spec.Components.RequestBodies[name]
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
	name, err := componentName(ref)
	if err != nil {
		return nil, err
	}

	if r, ok := g.responses[name]; ok {
		return r, nil
	}

	component, found := g.spec.Components.Responses[name]
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
