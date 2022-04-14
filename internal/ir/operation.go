package ir

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
)

type Operation struct {
	Name        string
	Description string
	PathParts   []*PathPart
	Params      []*Parameter
	Request     *Request
	Response    *Response
	Security    []*SecurityRequirement
	Spec        *openapi.Operation
}

func (op Operation) GoDoc() []string {
	return prettyDoc(op.Description)
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
	Spec *openapi.Parameter
}

// Default returns default value of this field, if it is set.
func (op Parameter) Default() Default {
	var schema *jsonschema.Schema
	if spec := op.Spec; spec != nil {
		schema = spec.Schema
	}
	if schema != nil {
		return Default{
			Value: schema.Default,
			Set:   schema.DefaultSet,
		}
	}
	if typ := op.Type; typ != nil {
		return typ.Default()
	}
	return Default{}
}

func (op Parameter) GoDoc() []string {
	if op.Spec == nil {
		return nil
	}
	return prettyDoc(op.Spec.Description)
}

type Request struct {
	Type     *Type
	Contents map[ContentType]*Type
	Spec     *openapi.RequestBody
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

func (t ContentType) String() string { return string(t) }

func (t ContentType) JSON() bool { return t == ContentTypeJSON }

func (t ContentType) OctetStream() bool { return t == ContentTypeOctetStream }

func (t ContentType) EncodedDataTypeGo() string {
	switch t {
	case ContentTypeJSON:
		return "*jx.Encoder"
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
}

type StatusResponse struct {
	Wrapped   bool
	NoContent *Type
	Contents  map[ContentType]*Type
	Spec      *openapi.Response
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
