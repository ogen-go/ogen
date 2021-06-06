package gen

import (
	"github.com/ernado/ogen"
)

type Generator struct {
	spec       *ogen.Spec
	components []componentStructDef
	groups     []pathGroupDef
	server     serverDef
}

func NewGenerator(spec *ogen.Spec) (*Generator, error) {
	g := &Generator{
		spec: spec,
	}

	if err := g.generateComponents(); err != nil {
		return nil, err
	}

	if err := g.generateServer(); err != nil {
		return nil, err
	}

	if err := g.generatePaths(); err != nil {
		return nil, err
	}

	return g, nil
}
