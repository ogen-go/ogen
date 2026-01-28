package ir

import (
	"fmt"
	"net/textproto"
	"slices"
	"strconv"
	"strings"

	"github.com/ogen-go/ogen/internal/xmaps"
)

type ResponseInfo struct {
	Type           *Type
	Encoding       Encoding
	ContentType    ContentType
	StatusCode     int
	NoContent      bool
	WithStatusCode bool
	WithHeaders    bool
	JSONStreaming  bool
	RawResponse    bool
	OpenTelemetry  bool
	Headers        map[string]*Parameter
}

func (r ResponseInfo) ContentTypeHeader() string {
	switch r.ContentType {
	case "application/json", "text/html", "text/plain":
		return fmt.Sprintf(`"%s; charset=utf-8"`, r.ContentType)
	default:
		return fmt.Sprintf(`%q`, r.ContentType)
	}
}

var corsSimpleResponseHeaders = map[string]struct{}{
	"Cache-Control":    {},
	"Content-Language": {},
	"Content-Length":   {},
	"Content-Type":     {},
	"Expires":          {},
	"Last-Modified":    {},
	"Pragma":           {},
}

func (r ResponseInfo) ExposeHeadersHeader() string {
	var hdr strings.Builder
	for _, header := range xmaps.SortedKeys(r.Headers) {
		header := textproto.CanonicalMIMEHeaderKey(header)
		if _, ok := corsSimpleResponseHeaders[header]; ok {
			continue
		}
		if hdr.Len() != 0 {
			hdr.WriteByte(',')
		}
		hdr.WriteString(header)
	}
	if hdr.Len() == 0 {
		return ""
	}
	return strconv.Quote(hdr.String())
}

func sortResponseInfos(result []ResponseInfo) {
	slices.SortStableFunc(result, func(l, r ResponseInfo) int {
		// Default responses has zero status code.
		if l.WithStatusCode {
			l.StatusCode = 999
		}
		if r.WithStatusCode {
			r.StatusCode = 999
		}
		if l.StatusCode != r.StatusCode {
			return l.StatusCode - r.StatusCode
		}
		if cmp := strings.Compare(l.ContentType.String(), r.ContentType.String()); cmp != 0 {
			return cmp
		}
		return strings.Compare(l.Type.Name, r.Type.Name)
	})
}

func (op *Operation) ListResponseTypes(otel bool) []ResponseInfo {
	var result []ResponseInfo
	for statusCode, resp := range op.Responses.StatusCode {
		if noc := resp.NoContent; noc != nil {
			result = append(result, ResponseInfo{
				Type:           noc,
				StatusCode:     statusCode,
				NoContent:      true,
				WithStatusCode: resp.WithStatusCode,
				WithHeaders:    resp.WithHeaders,
				OpenTelemetry:  otel,
				Headers:        resp.Headers,
			})
			continue
		}
		for contentType, media := range resp.Contents {
			result = append(result, ResponseInfo{
				Type:           media.Type,
				Encoding:       media.Encoding,
				ContentType:    contentType,
				StatusCode:     statusCode,
				WithStatusCode: resp.WithStatusCode,
				WithHeaders:    resp.WithHeaders,
				JSONStreaming:  media.JSONStreaming,
				RawResponse:    media.RawResponse,
				OpenTelemetry:  otel,
				Headers:        resp.Headers,
			})
		}
	}

	for _, resp := range op.Responses.Pattern {
		if resp == nil {
			continue
		}

		if noc := resp.NoContent; noc != nil {
			result = append(result, ResponseInfo{
				Type:           noc,
				NoContent:      true,
				WithStatusCode: resp.WithStatusCode,
				WithHeaders:    resp.WithHeaders,
				OpenTelemetry:  otel,
				Headers:        resp.Headers,
			})
			continue
		}
		for contentType, media := range resp.Contents {
			result = append(result, ResponseInfo{
				Type:           media.Type,
				Encoding:       media.Encoding,
				ContentType:    contentType,
				WithStatusCode: resp.WithStatusCode,
				WithHeaders:    resp.WithHeaders,
				JSONStreaming:  media.JSONStreaming,
				RawResponse:    media.RawResponse,
				OpenTelemetry:  otel,
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
				OpenTelemetry:  otel,
				Headers:        def.Headers,
			})
		}
		for contentType, media := range def.Contents {
			result = append(result, ResponseInfo{
				Type:           media.Type,
				Encoding:       media.Encoding,
				ContentType:    contentType,
				WithStatusCode: def.WithStatusCode,
				WithHeaders:    def.WithHeaders,
				JSONStreaming:  media.JSONStreaming,
				RawResponse:    media.RawResponse,
				OpenTelemetry:  otel,
				Headers:        def.Headers,
			})
		}
	}

	sortResponseInfos(result)
	return result
}
