package gen

import (
	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/validator"
)

type Generator struct {
	spec    *ogen.Spec
	methods []*Method

	schemas       map[string]*Schema
	requestBodies map[string]*RequestBody
	responses     map[string]*Response
	interfaces    map[string]*Interface
}

func NewGenerator(spec *ogen.Spec) (*Generator, error) {
	if err := validator.Validate(spec); err != nil {
		return nil, err
	}

	g := &Generator{
		spec:          spec,
		schemas:       map[string]*Schema{},
		requestBodies: map[string]*RequestBody{},
		responses:     map[string]*Response{},
		interfaces:    map[string]*Interface{},
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
