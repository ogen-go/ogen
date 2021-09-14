package gen

import (
	"fmt"
	"strings"

	"github.com/ogen-go/ogen"
)

// componentsParameter searches parameter defined in components section.
func (g *Generator) componentsParameter(ref string) (ogen.Parameter, bool) {
	if !strings.HasPrefix(ref, "#/components/parameters/") {
		return ogen.Parameter{}, false
	}

	targetName := strings.TrimPrefix(ref, "#/components/parameters/")
	for name, param := range g.spec.Components.Parameters {
		if name == targetName && param.Ref == "" {
			return param, true
		}
	}

	return ogen.Parameter{}, false
}

// componentsRequestBody searches requestBody defined in components section.
func (g *Generator) componentsRequestBody(ref string) (ogen.RequestBody, bool) {
	if !strings.HasPrefix(ref, "#/components/requestBodies/") {
		return ogen.RequestBody{}, false
	}

	targetName := strings.TrimPrefix(ref, "#/components/requestBodies/")
	for name, body := range g.spec.Components.RequestBodies {
		if name == targetName && body.Ref == "" {
			return body, true
		}
	}

	return ogen.RequestBody{}, false
}

// componentsResponse searches response defined in components section.
func (g *Generator) componentsResponse(ref string) (ogen.Response, bool) {
	if !strings.HasPrefix(ref, "#/components/responses/") {
		return ogen.Response{}, false
	}

	targetName := strings.TrimPrefix(ref, "#/components/responses/")
	for name, resp := range g.spec.Components.Responses {
		if name == targetName && resp.Ref == "" {
			return resp, true
		}
	}

	return ogen.Response{}, false
}

// componentsResponse searches response defined in components section.
func (g *Generator) componentsSchema(ref string) (ogen.Schema, bool) {
	if !strings.HasPrefix(ref, "#/components/schemas/") {
		return ogen.Schema{}, false
	}

	targetName := strings.TrimPrefix(ref, "#/components/schemas/")
	for name, schema := range g.spec.Components.Schemas {
		if name == targetName && schema.Ref == "" {
			return schema, true
		}
	}

	return ogen.Schema{}, false
}

func componentRefGotype(ref string) (string, error) {
	if !strings.HasPrefix(ref, "#/components/schemas/") {
		return "", fmt.Errorf("invalid component reference: %s", ref)
	}

	targetName := strings.TrimPrefix(ref, "#/components/schemas/")
	return pascal(targetName), nil
}

func requestBodyRefGotype(ref string) (string, error) {
	if !strings.HasPrefix(ref, "#/components/requestBodies/") {
		return "", fmt.Errorf("invalid requestBody reference: %s", ref)
	}

	targetName := strings.TrimPrefix(ref, "#/components/requestBodies/")
	return pascal(targetName), nil
}

func responseRefGotype(ref string) (string, error) {
	if !strings.HasPrefix(ref, "#/components/responses/") {
		return "", fmt.Errorf("invalid responses reference: %s", ref)
	}

	targetName := strings.TrimPrefix(ref, "#/components/responses/")
	return pascal(targetName), nil
}
