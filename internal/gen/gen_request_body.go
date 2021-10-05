package gen

import (
	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/ast"
)

func (g *Generator) generateRequestBody(name string, body *ogen.RequestBody) (*ast.RequestBody, error) {
	if ref := body.Ref; ref != "" {
		return g.resolveRequestBody(ref)
	}

	rbody := ast.CreateRequestBody()
	// Iterate through request body contents...
	for contentType, media := range body.Content {
		// Referenced schema.
		if ref := media.Schema.Ref; ref != "" {
			schema, err := g.resolveSchema(ref)
			if err != nil {
				if xerrors.Is(err, errSkipSchema) {
					continue
				}
				return nil, err
			}

			rbody.Contents[contentType] = schema
			continue
		}

		// Inlined schema.
		// Create unique name based on method name and contentType.
		schemaName := pascal(name, contentType, "Request")
		schema, err := g.generateSchema(name, media.Schema)
		if xerrors.Is(err, errSkipSchema) {
			continue
		}
		if err != nil {
			return nil, xerrors.Errorf("content: %s: parse schema: %w", contentType, err)
		}

		if schema.Is(ast.KindPrimitive, ast.KindArray) {
			schema = ast.CreateSchemaAlias(schemaName, schema.Type())
		}

		g.schemas[schema.Name] = schema
		rbody.Contents[contentType] = schema
	}

	return rbody, nil
}
