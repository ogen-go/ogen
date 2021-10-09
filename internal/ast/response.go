package ast

import (
	"sort"
	"strings"
)

type MethodResponse struct {
	StatusCode map[int]*Response
	Default    *Response
}

func CreateMethodResponses() *MethodResponse {
	return &MethodResponse{StatusCode: map[int]*Response{}}
}

type Response struct {
	NoContent *Schema
	Contents  map[string]*Schema
}

func (r *Response) Implement(iface *Interface) {
	if s := r.NoContent; s != nil {
		s.Implement(iface)
	}

	for _, schema := range r.Contents {
		schema.Implement(iface)
	}
}

func (r *Response) Unimplement(iface *Interface) {
	if s := r.NoContent; s != nil {
		s.Unimplement(iface)
	}

	for _, schema := range r.Contents {
		schema.Unimplement(iface)
	}
}

type ResponseInfo struct {
	Schema      *Schema
	StatusCode  int
	Default     bool
	ContentType string
	NoContent   bool
}

func (m *Method) ListResponseSchemas() []ResponseInfo {
	var result []ResponseInfo
	for statusCode, resp := range m.Responses.StatusCode {
		if resp.NoContent != nil {
			result = append(result, ResponseInfo{
				Schema:     resp.NoContent,
				StatusCode: statusCode,
				NoContent:  true,
			})
			continue
		}
		for contentType, schema := range resp.Contents {
			result = append(result, ResponseInfo{
				Schema:      schema,
				StatusCode:  statusCode,
				ContentType: contentType,
			})
		}
	}

	if def := m.Responses.Default; def != nil {
		if noc := def.NoContent; noc != nil {
			result = append(result, ResponseInfo{
				Schema:    noc,
				Default:   true,
				NoContent: true,
			})
		}
		for contentType, schema := range def.Contents {
			result = append(result, ResponseInfo{
				Schema:      schema,
				Default:     true,
				ContentType: contentType,
			})
		}
	}

	sort.SliceStable(result, func(i, j int) bool {
		l, r := result[i], result[j]
		// Default responses has zero status code.
		if l.Default {
			l.StatusCode = 999
		}
		if r.Default {
			r.StatusCode = 999
		}
		if l.StatusCode != r.StatusCode {
			return l.StatusCode < r.StatusCode
		}
		return strings.Compare(l.ContentType, r.ContentType) < 0
	})

	return result
}
