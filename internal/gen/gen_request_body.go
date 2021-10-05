package gen

import (
	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/ast"
)

func (g *Generator) generateRequestBody(methodName string, body *ogen.RequestBody) (*ast.RequestBody, error) {
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

	rbody := ast.CreateRequestBody()
	// Iterate through request body contents...
	for contentType, media := range body.Content {
		// Referenced schema.
		if ref := media.Schema.Ref; ref != "" {
			schema, err := g.resolveSchema(ref)
			if err != nil {
				return nil, err
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

		if schema.Is(ast.KindPrimitive, ast.KindArray) {
			schema = ast.CreateSchemaAlias(name, schema.Type())
		}

		g.schemas[schema.Name] = schema
		rbody.Contents[contentType] = schema
	}

	return rbody, nil
}
