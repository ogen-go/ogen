package gen

import (
	"fmt"
	"strings"

	"github.com/ogen-go/ogen"
)

const openapiVersion = "3.0.3"

type Generator struct {
	spec       *ogen.Spec
	components []componentStructDef
	server     serverDef
}

func NewGenerator(spec *ogen.Spec) (*Generator, error) {
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

	if err := g.generateComponents(); err != nil {
		return nil, err
	}

	if err := g.generateServer(); err != nil {
		return nil, err
	}

	return g, nil
}
