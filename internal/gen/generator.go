package gen

import (
	"github.com/ernado/ogen"
)

type Generator struct {
	spec *ogen.Spec
}

func NewGenerator(spec *ogen.Spec) (*Generator, error) {
	g := &Generator{
		spec: spec,
	}

	return g, nil
}
