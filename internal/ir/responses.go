package ir

import (
	"sort"
	"strings"
)

type ResponseInfo struct {
	Type        *Type
	StatusCode  int
	Default     bool
	ContentType ContentType
	NoContent   bool
}

func (op *Operation) ListResponseTypes() []ResponseInfo {
	var result []ResponseInfo
	for statusCode, resp := range op.Response.StatusCode {
		if resp.NoContent != nil {
			result = append(result, ResponseInfo{
				Type:       resp.NoContent,
				StatusCode: statusCode,
				NoContent:  true,
			})
			continue
		}
		for contentType, typ := range resp.Contents {
			result = append(result, ResponseInfo{
				Type:        typ,
				StatusCode:  statusCode,
				ContentType: contentType,
			})
		}
	}

	if def := op.Response.Default; def != nil {
		if noc := def.NoContent; noc != nil {
			result = append(result, ResponseInfo{
				Type:      noc,
				Default:   true,
				NoContent: true,
			})
		}
		for contentType, typ := range def.Contents {
			result = append(result, ResponseInfo{
				Type:        typ,
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
		return strings.Compare(string(l.ContentType), string(r.ContentType)) < 0
	})

	return result
}
