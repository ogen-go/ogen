package gen

import (
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
