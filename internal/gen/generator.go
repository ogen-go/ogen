package gen

import (
	"fmt"
	"strings"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/validator"
)

const openapiVersion = "3.0.3"

type Generator struct {
	spec    *ogen.Spec
	schemas []schemaStructDef
	server  serverDef
}

func NewGenerator(spec *ogen.Spec) (*Generator, error) {
	if err := validator.Validate(spec); err != nil {
		return nil, err
	}
	initSpec(spec)
	g := &Generator{
		spec: spec,
	}

	if strings.TrimSpace(spec.OpenAPI) == "" {
		return nil, fmt.Errorf("openapi version is not defined")
	}

	if spec.OpenAPI != openapiVersion {
		return nil, fmt.Errorf(
			"unsupported OpenAPI version: %s (expected: %s)",
			spec.OpenAPI,
			openapiVersion,
		)
	}

	if err := g.generateSchemaComponents(); err != nil {
		return nil, err
	}

	if err := g.generateServer(); err != nil {
		return nil, err
	}

	return g, nil
}

func initSpec(spec *ogen.Spec) {
	if spec.Components == nil {
		spec.Components = &ogen.Components{}
	}
	if spec.Components.Parameters == nil {
		spec.Components.Parameters = make(map[string]ogen.Parameter)
	}
	if spec.Components.RequestBodies == nil {
		spec.Components.RequestBodies = make(map[string]ogen.RequestBody)
	}
	if spec.Components.Responses == nil {
		spec.Components.Responses = make(map[string]ogen.Response)
	}
	if spec.Components.Schemas == nil {
		spec.Components.Schemas = make(map[string]ogen.Schema)
	}
}
