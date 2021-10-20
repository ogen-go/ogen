package gen

import (
	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/ast"
)

func (g *Generator) generateRequestBody(name string, body *ogen.RequestBody) (*ast.RequestBody, error) {
	if ref := body.Ref; ref != "" {
		rbody, err := g.resolveRequestBody(ref)
		if err != nil {
			return nil, xerrors.Errorf("resolve '%s' reference: %w", ref, err)
		}

		return rbody, nil
	}

	reqBody := ast.CreateRequestBody()
	reqBody.Required = body.Required

	// Iterate through request body contents...
	for contentType, media := range body.Content {
		schemaName := pascal(name, contentType, "Request")
		schema, err := g.generateSchema(schemaName, media.Schema)
		if err != nil {
			return nil, xerrors.Errorf("content: %s: parse schema: %w", contentType, err)
		}

		if inlined := media.Schema.Ref == ""; inlined {
			// Wrap scalar type with an alias.
			// It is necessary because schema should satisfy
			// <methodName>Request interface.
			//
			// Alias can be removed later in the simplification stage
			// if there's no other requests.
			if schema.Is(ast.KindPrimitive, ast.KindArray, ast.KindPointer) {
				schema = ast.Alias(schemaName, schema)
			}

			// Register schema.
			g.schemas[schema.Name] = schema
		}

		reqBody.Contents[contentType] = schema
	}

	return reqBody, nil
}
