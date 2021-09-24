package gen

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ogen-go/ogen"
)

type methodResponse struct {
	Responses map[int]*Response
	Default   *Response
}

func (g *Generator) generateResponses(methodName string, methodResponses ogen.Responses) (*methodResponse, error) {
	var (
		responses   = make(map[int]*Response)
		defaultResp *Response
	)

	// Iterate over method responses...
	for status, responseSchema := range methodResponses {
		// Default response.
		if status == "default" {
			// Referenced response.
			if ref := responseSchema.Ref; ref != "" {
				// Validate reference & get response component name.
				name, err := componentName(ref)
				if err != nil {
					return nil, err
				}

				// Lookup for alias response.
				if alias, ok := g.responses[name+"Default"]; ok {
					defaultResp = alias
					continue
				}

				// Lookup for reference response.
				response, found := g.responses[name]
				if !found {
					return nil, fmt.Errorf("response by reference '%s', not found", ref)
				}

				aliasResponse := g.createResponse(name + "Default")
				for contentType, schema := range response.Contents {
					aliasResponse.Contents[contentType] = g.wrapStatusCode(schema)
				}
				if schema := response.NoContent; schema != nil {
					response.NoContent = g.wrapStatusCode(schema)
				}

				defaultResp = aliasResponse
				continue
			}

			// Inlined response.
			// Use method name + Default as prefix for response schemas.
			response, err := g.generateResponse(methodName+"Default", responseSchema)
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

			defaultResp = response
			continue
		}

		statusCode, err := strconv.Atoi(status)
		if err != nil {
			return nil, fmt.Errorf("invalid status code: '%s'", status)
		}

		// Referenced response.
		if ref := responseSchema.Ref; ref != "" {
			// Validate reference & get response component name.
			name, err := componentName(ref)
			if err != nil {
				return nil, err
			}

			// Lookup for response component.
			componentResponse, found := g.responses[name]
			if !found {
				return nil, fmt.Errorf("response by reference '%s' not found", ref)
			}

			responses[statusCode] = componentResponse
			continue
		}

		responseName := methodName
		if len(responses) > 1 {
			// Use status code in response name to avoid collisions.
			responseName = methodName + http.StatusText(statusCode)
		}

		resp, err := g.generateResponse(responseName, responseSchema)
		if err != nil {
			return nil, fmt.Errorf("invalid status code: '%s'", status)
		}

		responses[statusCode] = resp
	}

	return &methodResponse{
		Responses: responses,
		Default:   defaultResp,
	}, nil
}

func (g *Generator) generateResponse(name string, resp ogen.Response) (*Response, error) {
	response := g.createResponse(name)

	// Response without content.
	// Create empty struct.
	if len(resp.Content) == 0 {
		s := g.createSchemaSimple(name, "struct{}")
		g.schemas[s.Name] = s
		response.NoContent = s
		return response, nil
	}

	for contentType, media := range resp.Content {
		// Create unique response name.
		responseStructName := name + "Response"
		if len(resp.Content) > 1 {
			responseStructName = pascal(
				name+"_"+strings.ReplaceAll(contentType, "/", "_"),
			) + "Response"
		}

		// Referenced response schema.
		if ref := media.Schema.Ref; ref != "" {
			refSchemaName, err := componentName(ref)
			if err != nil {
				return nil, err
			}

			schema, found := g.schemas[refSchemaName]
			if !found {
				return nil, fmt.Errorf("schema referenced by '%s' not found", ref)
			}

			// Response have only one content.
			// Use schema directly without creating new one.
			if len(resp.Content) == 1 {
				response.Contents[contentType] = schema
				continue
			}

			// Response have multiple contents.
			// Alias them with new response type.
			s := g.createSchemaSimple(responseStructName, schema.Name)
			g.schemas[s.Name] = s
			response.Contents[contentType] = s
			continue
		}

		// Inlined response schema.
		s, err := g.generateSchema(responseStructName, media.Schema)
		if err != nil {
			return nil, fmt.Errorf("content: %s: parse schema: %w", contentType, err)
		}

		g.schemas[s.Name] = s
		response.Contents[contentType] = s
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
func (g *Generator) wrapStatusCode(schema *Schema) *Schema {
	// Use 'StatusCode' postfix for wrapper struct name
	// to avoid name collision with original response schema.
	newSchema := g.createSchemaStruct(schema.Name + "StatusCode")
	newSchema.Fields = []SchemaField{
		{
			Name: "StatusCode",
			Tag:  "-",
			Type: "int",
		},
		{
			Name: "Response",
			Tag:  "-",
			Type: schema.typeName(),
		},
	}
	g.schemas[newSchema.Name] = newSchema
	return newSchema
}
