package ir

import (
	"sort"
)

type ResponseInfo struct {
	Type           *Type
	StatusCode     int
	ContentType    ContentType
	NoContent      bool
	WithStatusCode bool
	WithHeaders    bool
	Headers        map[string]*Parameter
}

func (op *Operation) ListResponseTypes() []ResponseInfo {
	var result []ResponseInfo
	for statusCode, resp := range op.Responses.StatusCode {
		if noc := resp.NoContent; noc != nil {
			result = append(result, ResponseInfo{
				Type:           noc,
				StatusCode:     statusCode,
				NoContent:      true,
				WithStatusCode: resp.WithStatusCode,
				WithHeaders:    resp.WithHeaders,
				Headers:        resp.Headers,
			})
			continue
		}
		for contentType, typ := range resp.Contents {
			result = append(result, ResponseInfo{
				Type:           typ,
				StatusCode:     statusCode,
				ContentType:    contentType,
				WithStatusCode: resp.WithStatusCode,
				WithHeaders:    resp.WithHeaders,
				Headers:        resp.Headers,
			})
		}
	}

	if def := op.Responses.Default; def != nil {
		if noc := def.NoContent; noc != nil {
			result = append(result, ResponseInfo{
				Type:           noc,
				NoContent:      true,
				WithStatusCode: def.WithStatusCode,
				WithHeaders:    def.WithHeaders,
				Headers:        def.Headers,
			})
		}
		for contentType, typ := range def.Contents {
			result = append(result, ResponseInfo{
				Type:           typ,
				ContentType:    contentType,
				WithStatusCode: def.WithStatusCode,
				WithHeaders:    def.WithHeaders,
				Headers:        def.Headers,
			})
		}
	}

	sort.SliceStable(result, func(i, j int) bool {
		l, r := result[i], result[j]
		// Default responses has zero status code.
		if l.WithStatusCode {
			l.StatusCode = 999
		}
		if r.WithStatusCode {
			r.StatusCode = 999
		}
		if l.StatusCode != r.StatusCode {
			return l.StatusCode < r.StatusCode
		}
		return string(l.ContentType) < string(r.ContentType)
	})

	return result
}
