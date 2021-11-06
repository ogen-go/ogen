package ir

import (
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
	NoContent *Type
	Contents  map[ContentType]*Type
	Spec      *oas.Response
}
