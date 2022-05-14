package ir

import (
	"sort"
)

type ResponseInfo struct {
	Type        *Type
	StatusCode  int
	Default     bool
	ContentType ContentType
	NoContent   bool
	Wrapped     bool
}

func (op *Operation) ListResponseTypes() []ResponseInfo {
	var result []ResponseInfo
	for statusCode, resp := range op.Response.StatusCode {
		if noc := resp.NoContent; noc != nil {
			result = append(result, ResponseInfo{
				Type:       noc,
				StatusCode: statusCode,
				NoContent:  true,
				Wrapped:    resp.Wrapped,
			})
			continue
		}
		for contentType, typ := range resp.Contents {
			result = append(result, ResponseInfo{
				Type:        typ,
				StatusCode:  statusCode,
				ContentType: contentType,
				Wrapped:     resp.Wrapped,
			})
		}
	}

	if def := op.Response.Default; def != nil {
		if noc := def.NoContent; noc != nil {
			result = append(result, ResponseInfo{
				Type:      noc,
				Default:   true,
				NoContent: true,
				Wrapped:   def.Wrapped,
			})
		}
		for contentType, typ := range def.Contents {
			result = append(result, ResponseInfo{
				Type:        typ,
				Default:     true,
				ContentType: contentType,
				Wrapped:     def.Wrapped,
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
		return string(l.ContentType) < string(r.ContentType)
	})

	return result
}
