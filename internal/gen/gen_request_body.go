package gen

import (
	"fmt"
	"strings"

	"github.com/ogen-go/ogen"
)

func (g *Generator) generateRequestBody(methodName string, body *ogen.RequestBody) (*RequestBody, error) {
	if body.Ref != "" {
		// Convert requestBody reference to go type name.
		name, err := componentName(body.Ref)
		if err != nil {
			return nil, err
		}

		// Lookup for requestBody.
		rbody, found := g.requestBodies[name]
		if !found {
			return nil, fmt.Errorf("requestBody by reference '%s' not found", body.Ref)
		}

		return rbody, nil
	}

	rbody := g.createRequestBody()
	// Iterate through request body contents...
	for contentType, media := range body.Content {
		// Referenced schema.
		if ref := media.Schema.Ref; ref != "" {
			name, err := componentName(ref)
			if err != nil {
				return nil, err
			}

			// Lookup for schema.
			schema, found := g.schemas[name]
			if !found {
				return nil, fmt.Errorf("schema by reference '%s' not found", ref)
			}

			rbody.Contents[contentType] = schema
			continue
		}

		// Inlined schema.
		// Create unique name based on method name and contentType.
		inputName := methodName + "_" + strings.ReplaceAll(contentType, "/", "_") + "_Request"
		inputName = pascal(inputName)

		// Generate schema.
		schema, err := g.generateSchema(inputName, media.Schema)
		if err != nil {
			return nil, fmt.Errorf("content: %s: parse schema: %w", contentType, err)
		}

		// Register generated schema.
		g.schemas[inputName] = schema
		rbody.Contents[contentType] = schema
	}

	return rbody, nil
}
