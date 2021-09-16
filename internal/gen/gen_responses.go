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

				// Create new type containing referenced response
				// and status code field.
				referencedResponse, found := g.responses[name]
				if !found {
					return nil, fmt.Errorf("response by reference '%s', not found", ref)
				}

				aliasResponse := g.createResponse(name + "Default")
				for contentType, schema := range referencedResponse.Contents {
					alias := g.createSchema(schema.Name + "Default")
					alias.Simple = schema.Simple
					alias.Fields = append([]SchemaField{
						{
							Name: "StatusCode",
							Tag:  "-",
							Type: "int",
						},
					}, schema.Fields...)
					aliasResponse.Contents[contentType] = alias
				}

				if schema := referencedResponse.NoContent; schema != nil {
					alias := g.createSchema(schema.Name + "Default")
					alias.Simple = schema.Simple
					alias.Fields = append([]SchemaField{
						{
							Name: "StatusCode",
							Tag:  "-",
							Type: "int",
						},
					}, schema.Fields...)
					aliasResponse.NoContent = alias
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

			// Append status code for all response schemas fields.
			for _, schema := range response.Contents {
				schema.Fields = append([]SchemaField{
					{
						Name: "StatusCode",
						Tag:  "-",
						Type: "int",
					},
				}, schema.Fields...)
			}

			if response.NoContent != nil {
				response.NoContent.Fields = append([]SchemaField{
					{
						Name: "StatusCode",
						Tag:  "-",
						Type: "int",
					},
				}, response.NoContent.Fields...)
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
		response.NoContent = g.createSchemaSimple(name, "struct{}")
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
			response.Contents[contentType] = s
			continue
		}

		// Inlined response schema.
		schema, err := g.generateSchema(responseStructName, media.Schema)
		if err != nil {
			return nil, fmt.Errorf("content: %s: parse schema: %w", contentType, err)
		}

		g.schemas[responseStructName] = schema
		response.Contents[contentType] = schema
	}

	return response, nil
}
