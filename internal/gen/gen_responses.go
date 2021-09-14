package gen

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ogen-go/ogen"
)

func (g *Generator) generateResponses(name string, responses ogen.Responses) (map[int]*Response, error) {
	resps := make(map[int]*Response)
	for status, resp := range responses {
		if status == "default" {
			return nil, fmt.Errorf("default status code not supported")
		}

		statusCode, err := strconv.Atoi(status)
		if err != nil {
			return nil, fmt.Errorf("invalid status code: '%s'", status)
		}

		if ref := resp.Ref; ref != "" {
			typeName, err := responseRefGotype(ref)
			if err != nil {
				return nil, err
			}

			componentResponse, found := g.responses[typeName]
			if !found {
				panic("unreachable")
			}

			resps[statusCode] = componentResponse
			continue
		}

		respName := name
		if len(responses) > 1 {
			// Use status code in response name to avoid collisions.
			respName = name + http.StatusText(statusCode)
		}

		resp, err := g.generateResponse(respName, resp)
		if err != nil {
			return nil, fmt.Errorf("invalid status code: '%s'", status)
		}

		resps[statusCode] = resp
	}

	return resps, nil
}

func (g *Generator) generateResponse(name string, resp ogen.Response) (*Response, error) {
	response := &Response{
		Contents: map[string]*Schema{},
	}

	// Response without content.
	// Create empty struct.
	if len(resp.Content) == 0 {
		s := &Schema{
			Name:       name,
			Simple:     "struct{}",
			Implements: map[string]struct{}{},
		}
		response.NoContent = s
		g.schemas[name] = s
		return response, nil
	}

	for contentType, media := range resp.Content {
		respName := name + "Response"

		if len(resp.Content) > 1 {
			// Use content type in response name to avoid collisions.
			respName = pascal(
				name+"_"+strings.ReplaceAll(contentType, "/", "_"),
			) + "Response"
		}

		if ref := media.Schema.Ref; ref != "" {
			typeName, err := componentRefGotype(ref)
			if err != nil {
				return nil, err
			}

			schema, found := g.schemas[typeName]
			if !found {
				panic("unreachable")
			}

			// Response have only one content.
			// Use schema directly without creating new one.
			if len(resp.Content) == 1 {
				response.Contents[contentType] = schema
				continue
			}

			// Response have multiple contents.
			// Alias them with new response type.
			s := &Schema{
				Name:       respName,
				Simple:     schema.Name,
				Implements: map[string]struct{}{},
			}

			g.schemas[respName] = s
			response.Contents[contentType] = s
			continue
		}

		schema, err := g.generateSchema(respName, media.Schema)
		if err != nil {
			return nil, fmt.Errorf("content: %s: parse schema: %w", contentType, err)
		}

		g.schemas[respName] = schema
		response.Contents[contentType] = schema
	}

	return response, nil
}
