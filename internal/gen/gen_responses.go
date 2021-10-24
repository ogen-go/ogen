package gen

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/ast"
)

func (g *Generator) generateResponses(methodName string, responses ogen.Responses) (*ast.MethodResponse, error) {
	result := ast.CreateMethodResponses()
	if len(responses) == 0 {
		return nil, fmt.Errorf("no responses")
	}

	var statuses []string
	for stat := range responses {
		statuses = append(statuses, stat)
	}
	sort.Strings(statuses)

	// Iterate over method responses...
	for _, status := range statuses {
		response := responses[status]
		// Default response.
		if status == "default" {
			resp, err := g.createDefaultResponse(methodName, response)
			if err != nil {
				return nil, xerrors.Errorf("default: %w", err)
			}

			result.Default = resp
			continue
		}

		statusCode, err := strconv.Atoi(status)
		if err != nil {
			return nil, xerrors.Errorf("invalid status code: '%s'", status)
		}

		responseName := pascal(methodName)
		if len(responses) > 1 {
			// Avoid collision with <methodName>Response interface.
			responseName = pascal(methodName, http.StatusText(statusCode))
		}

		resp, err := g.generateResponse(responseName, response)
		if err != nil {
			return nil, xerrors.Errorf("%s: %w", status, err)
		}

		result.StatusCode[statusCode] = resp
	}

	return result, nil
}

// createDefaultResponse creates new default response.
func (g *Generator) createDefaultResponse(methodName string, r ogen.Response) (*ast.Response, error) {
	if ref := r.Ref; ref != "" {
		// Lookup for reference response.
		response, err := g.resolveResponse(ref)
		if err != nil {
			return nil, err
		}

		alias := ast.CreateResponse()
		for contentType, schema := range response.Contents {
			alias.Contents[contentType] = g.wrapStatusCode(schema)
		}
		if schema := response.NoContent; schema != nil {
			response.NoContent = g.wrapStatusCode(schema)
		}

		return alias, nil
	}

	// Default response with no contents.
	if len(r.Content) == 0 {
		statusCode := ast.Struct(methodName + "Default")
		statusCode.Fields = append(statusCode.Fields, ast.SchemaField{
			Name: "StatusCode",
			Type: ast.Primitive(ast.Int),
			Tag:  "-",
		})
		g.schemas[methodName+"Default"] = statusCode
		return &ast.Response{NoContent: statusCode}, nil
	}

	// Inlined response.
	// Use method name + Default as prefix for response schemas.
	response, err := g.generateResponse(methodName+"Default", r)
	if err != nil {
		return nil, err
	}

	// We need to inject StatusCode field to response structs somehow...
	// Iterate over all responses and create new response schema wrapper:
	//
	// type <WrapperName> struct {
	//     StatusCode int            `json:"-"`
	//     Response   <ResponseType> `json:"-"`
	// }
	for contentType, schema := range response.Contents {
		defaultSchema := g.wrapStatusCode(schema)
		response.Contents[contentType] = defaultSchema
	}
	if schema := response.NoContent; schema != nil {
		response.NoContent = g.wrapStatusCode(schema)
	}

	return response, nil
}

// generateResponse creates new response based on schema definition.
func (g *Generator) generateResponse(respName string, resp ogen.Response) (*ast.Response, error) {
	if ref := resp.Ref; ref != "" {
		resp, err := g.resolveResponse(ref)
		if err != nil {
			return nil, xerrors.Errorf("resolve '%s' reference: %w", ref, err)
		}

		return resp, nil
	}

	response := ast.CreateResponse()

	// Response without content.
	// Create empty struct.
	if len(resp.Content) == 0 {
		s := ast.Alias(respName, ast.Primitive(ast.EmptyStruct))
		g.schemas[s.Name] = s
		response.NoContent = s
		return response, nil
	}

	var contentTypes []string
	for ct := range resp.Content {
		contentTypes = append(contentTypes, ct)
	}
	sort.Strings(contentTypes)

	for _, contentType := range contentTypes {
		media := resp.Content[contentType]
		// Create unique response name.
		name := respName
		if len(resp.Content) > 1 {
			name = pascal(respName, contentType)
		}

		schema, err := g.generateSchema(name, media.Schema)
		if err != nil {
			return nil, xerrors.Errorf("content: %s: schema: %w", contentType, err)
		}

		if inlined := media.Schema.Ref == ""; inlined {
			// Wrap scalar type with an alias.
			// It is necessary because schema should satisfy
			// <methodName>Response interface.
			//
			// Alias can be removed later in the simplification stage
			// if there's no other responses.
			if schema.Is(ast.KindPrimitive, ast.KindArray, ast.KindPointer) {
				schema = ast.Alias(name, schema)
			}

			// Register schema.
			g.schemas[schema.Name] = schema
		}

		response.Contents[contentType] = schema
	}

	return response, nil
}

// wrapStatusCode wraps provided schema with newtype containing StatusCode field.
//
// Example 1:
//   Schema:
//   type FoobarGetDefaultResponse {
//       Message string `json:"message"`
//       Code    int64  `json:"code"`
//   }
//
//   Wrapper:
//   type FoobarGetDefaultResponseStatusCode {
//       StatusCode int                      `json:"-"`
//       Response   FoobarGetDefaultResponse `json:"-"`
//   }
//
// Example 2:
//   Schema:
//   type FoobarGetDefaultResponse string
//
//   Wrapper:
//   type FoobarGetDefaultResponseStatusCode {
//       StatusCode int    `json:"-"`
//       Response   string `json:"-"`
//   }
//
// TODO: Remove unused schema (Example 2).
func (g *Generator) wrapStatusCode(schema *ast.Schema) *ast.Schema {
	if !schema.Is(ast.KindStruct, ast.KindAlias, ast.KindPointer) {
		panic("unreachable")
	}

	if s, ok := g.schemas[schema.Name+"StatusCode"]; ok {
		return s
	}

	// Use 'StatusCode' postfix for wrapper struct name
	// to avoid name collision with original response schema.
	newSchema := ast.Struct(schema.Name + "StatusCode")
	newSchema.Fields = []ast.SchemaField{
		{
			Name: "StatusCode",
			Type: ast.Primitive(ast.Int),
			Tag:  "-",
		},
		{
			Name: "Response",
			Type: schema,
			Tag:  "-",
		},
	}
	g.schemas[newSchema.Name] = newSchema
	return newSchema
}
