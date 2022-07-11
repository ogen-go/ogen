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
	Summary     string
	Description string
	Deprecated  bool
	PathParts   []*PathPart
	Params      []*Parameter
	Request     *Request
	Responses   *Responses
	Security    []*SecurityRequirement
	Spec        *openapi.Operation
}

func (op Operation) PrettyOperationID() string {
	s := op.Spec
	if id := s.OperationID; id != "" {
		return id
	}
	return strings.ToUpper(s.HTTPMethod) + " " + s.Path.String()
}

func (op Operation) GoDoc() []string {
	doc := op.Description
	if doc == "" {
		doc = op.Summary
	}

	var notice string
	if op.Deprecated {
		notice = "Deprecated: schema marks this operation as deprecated."
	}
	return prettyDoc(doc, notice)
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

func (op Parameter) GoDoc() []string {
	s := op.Spec
	if s == nil {
		return nil
	}

	var notice string
	if s.Deprecated {
		notice = "Deprecated: schema marks this parameter as deprecated."
	}
	return prettyDoc(s.Description, notice)
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

type Request struct {
	Type     *Type
	Contents map[ContentType]*Type
	Spec     *openapi.RequestBody
}

// ContentType of body.
type ContentType string

const (
	// ContentTypeJSON is ContentType for json.
	ContentTypeJSON ContentType = "application/json"
	// ContentTypeFormURLEncoded is ContentType for URL-encoded form.
	ContentTypeFormURLEncoded ContentType = "application/x-www-form-urlencoded"
	// ContentTypeMultipart is ContentType for multipart form.
	ContentTypeMultipart ContentType = "multipart/form-data"
	// ContentTypeOctetStream is ContentType for binary.
	ContentTypeOctetStream ContentType = "application/octet-stream"
	// ContentTypeTextPlain is ContentType for text.
	ContentTypeTextPlain ContentType = "text/plain"
)

func (t ContentType) String() string { return string(t) }

func (t ContentType) Mask() bool { return strings.ContainsRune(string(t), '*') }

func (t ContentType) JSON() bool { return t == ContentTypeJSON }

func (t ContentType) FormURLEncoded() bool { return t == ContentTypeFormURLEncoded }

func (t ContentType) MultipartForm() bool { return t == ContentTypeMultipart }

func (t ContentType) OctetStream() bool { return t == ContentTypeOctetStream }

func (t ContentType) TextPlain() bool { return t == ContentTypeTextPlain }

type Media struct {
	Encoding ContentType
	Type     *Type
}

type Responses struct {
	Type       *Type
	StatusCode map[int]*Response
	Default    *Response
}

func (r *Responses) BlankType() bool {
	if r.Default != nil || len(r.StatusCode) > 1 {
		return false
	}

	if len(r.StatusCode) == 0 {
		panic("unreachable")
	}

	for _, r := range r.StatusCode {
		tt := r.NoContent
		if tt != nil && tt.IsStruct() && len(tt.Fields) == 0 {
			return true
		}
	}

	return false
}

type Response struct {
	NoContent *Type
	Contents  map[ContentType]*Type
	Spec      *openapi.Response
	Headers   map[string]*Parameter

	// Indicates that all response types
	// are wrappers with StatusCode field.
	WithStatusCode bool

	// Indicates that all response types
	// are wrappers with response header fields.
	WithHeaders bool

	// Note that if NoContent is false
	// (i.e. response has specified contents)
	// and (WithStatusCode || WithHeaders) == true
	// all wrapper types will also have a Response field
	// which will contain the actual response body.
}

func (s Response) ResponseInfo() []ResponseInfo {
	var result []ResponseInfo

	if noc := s.NoContent; noc != nil {
		result = append(result, ResponseInfo{
			Type:           noc,
			NoContent:      true,
			WithStatusCode: s.WithStatusCode,
			WithHeaders:    s.WithHeaders,
		})
	}
	for contentType, typ := range s.Contents {
		result = append(result, ResponseInfo{
			Type:           typ,
			ContentType:    contentType,
			WithStatusCode: s.WithStatusCode,
			WithHeaders:    s.WithHeaders,
		})
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
