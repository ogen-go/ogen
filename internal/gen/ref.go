package gen

import "github.com/ogen-go/ogen"

func (g *Generator) resolveSchema(ref string) (*Schema, error) {
	name, err := componentName(ref)
	if err != nil {
		return nil, err
	}

	return g.generateSchema(name, ogen.Schema{
		Ref: ref,
	})
}
