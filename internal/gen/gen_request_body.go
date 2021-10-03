package gen

import (
	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen"
)

func (g *Generator) generateRequestBody(methodName string, body *ogen.RequestBody) (*RequestBody, error) {
	if body.Ref != "" {
		// Convert requestBody reference to go type name.
		name, err := componentName(body.Ref)
		if err != nil {
			return nil, xerrors.Errorf("name: %w", err)
		}

		// Lookup for requestBody.
		rbody, found := g.requestBodies[name]
		if !found {
			return nil, xerrors.Errorf("requestBody by reference '%s' not found", body.Ref)
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
				return nil, xerrors.Errorf("schema by reference '%s' not found", ref)
			}

			rbody.Contents[contentType] = schema
			continue
		}

		// Inlined schema.
		// Create unique name based on method name and contentType.
		name := pascal(methodName, contentType, "Request")
		schema, err := g.generateSchema(name, media.Schema)
		if xerrors.Is(err, errSkipSchema) {
			continue
		}
		if err != nil {
			return nil, xerrors.Errorf("content: %s: parse schema: %w", contentType, err)
		}

		// Register generated schema.
		g.schemas[name] = schema
		rbody.Contents[contentType] = schema
	}

	return rbody, nil
}
