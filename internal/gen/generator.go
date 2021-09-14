package gen

import (
	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/validator"
)

type Generator struct {
	spec    *ogen.Spec
	methods []*Method

	// Schemas parsed from components section.
	// map[GoType]*Schema
	schemas map[string]*Schema

	// RequestBodies parsed from components section.
	// map[GoType]*RequestBody
	requestBodies map[string]*RequestBody

	// map[IfaceName]map[Method]
	interfaces map[string]map[string]struct{}
}

func NewGenerator(spec *ogen.Spec) (*Generator, error) {
	if err := validator.Validate(spec); err != nil {
		return nil, err
	}

	g := &Generator{
		spec:          spec,
		schemas:       map[string]*Schema{},
		requestBodies: map[string]*RequestBody{},
		interfaces:    map[string]map[string]struct{}{},
	}

	if err := g.generateComponents(); err != nil {
		return nil, err
	}

	if err := g.generateMethods(); err != nil {
		return nil, err
	}

	g.simplify()
	return g, nil
}
