package gen

import (
	"fmt"
	"strings"

	"github.com/ogen-go/ogen"
)

func (g *Generator) generateRequestBody(name string, body *ogen.RequestBody) (*RequestBody, error) {
	if body.Ref != "" {
		typeName, err := requestBodyRefGotype(body.Ref)
		if err != nil {
			return nil, err
		}

		rbody, found := g.requestBodies[typeName]
		if !found {
			panic("unreachable")
		}

		return rbody, nil
	}

	rbody := &RequestBody{
		Contents: map[string]*Schema{},
		Required: body.Required,
	}
	for contentType, media := range body.Content {
		if ref := media.Schema.Ref; ref != "" {
			typeName, err := componentRefGotype(ref)
			if err != nil {
				return nil, err
			}

			schema, found := g.schemas[typeName]
			if !found {
				panic("unreachable")
			}

			rbody.Contents[contentType] = schema
			continue
		}

		inputName := name + "_" + strings.ReplaceAll(contentType, "/", "_") + "_Request"
		inputName = pascal(inputName)

		schema, err := g.generateSchema(inputName, media.Schema)
		if err != nil {
			return nil, fmt.Errorf("content: %s: parse schema: %w", contentType, err)
		}

		g.schemas[inputName] = schema
		rbody.Contents[contentType] = schema
	}

	return rbody, nil
}
