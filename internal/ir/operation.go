package ir

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ogen-go/ogen/internal/oas"
)

type Operation struct {
	Name      string
	PathParts []*PathPart
	Params    []*Parameter
	Request   *Request
	Response  *Response
	Spec      *oas.Operation
}

type PathPart struct {
	Raw   string
	Param *Parameter
}

func (p PathPart) String() string {
	if p.Param != nil {
		return fmt.Sprintf("{%s}", p.Param.Spec.Name)
	}
	return p.Raw
}

type Parameter struct {
	Name string
	Type *Type
	Spec *oas.Parameter
}

type Request struct {
	Type     *Type
	Contents map[ContentType]*Type
	Required bool
	Spec     *oas.RequestBody
}

type Content struct {
	ContentType ContentType
	Type        *Type
}

// ContentType of body.
type ContentType string

const (
	// ContentTypeJSON is ContentType for json.
	ContentTypeJSON ContentType = "application/json"
	// ContentTypeOctetStream is ContentType for binary.
	ContentTypeOctetStream ContentType = "application/octet-stream"
)

func (t ContentType) JSON() bool { return t == ContentTypeJSON }

func (t ContentType) OctetStream() bool { return t == ContentTypeOctetStream }

func (t ContentType) EncodedDataTypeGo() string {
	switch t {
	case ContentTypeJSON:
		return "*bytes.Buffer"
	case ContentTypeOctetStream:
		return "io.Reader"
	default:
		return "io.Reader"
	}
}

func (t ContentType) Name() string {
	switch t {
	case ContentTypeJSON:
		return "JSON"
	case ContentTypeOctetStream:
		return "OctetStream"
	default:
		return ""
	}
}

type Response struct {
	Type       *Type
	StatusCode map[int]*StatusResponse
	Default    *StatusResponse
	Spec       *oas.OperationResponse
}

type StatusResponse struct {
	Wrapped   bool
	NoContent *Type
	Contents  map[ContentType]*Type
	Spec      *oas.Response
}

func (s StatusResponse) ResponseInfo() []ResponseInfo {
	var result []ResponseInfo

	if noc := s.NoContent; noc != nil {
		result = append(result, ResponseInfo{
			Type:      noc,
			Default:   true,
			NoContent: true,
		})
	}
	for contentType, typ := range s.Contents {
		result = append(result, ResponseInfo{
			Type:        typ,
			Default:     true,
			ContentType: contentType,
		})
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
