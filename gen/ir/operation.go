package ir

import (
	"fmt"
	"strings"

	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
)

// WebhookInfo contains information about webhook.
type WebhookInfo struct {
	// Name is the name of the webhook.
	Name string
}

type Operation struct {
	Name           string
	Summary        string
	Description    string
	Deprecated     bool
	WebhookInfo    *WebhookInfo
	PathParts      []*PathPart
	Params         []*Parameter
	Request        *Request
	Responses      *Responses
	Security       SecurityRequirements
	Spec           *openapi.Operation
	OperationGroup string
}

type OperationGroup struct {
	Name       string
	Operations []*Operation
}

// OTELAttribute represents OpenTelemetry attribute defined by otelogen package.
type OTELAttribute struct {
	// Key is a name of the attribute constructor in otelogen package.
	Key string
	// Value is a value of the attribute.
	Value string
}

// String returns call to the constructor of this attribute.
func (a OTELAttribute) String() string {
	return fmt.Sprintf("otelogen.%s(%q)", a.Key, a.Value)
}

// OTELAttributes returns OpenTelemetry attributes for this operation.
func (op Operation) OTELAttributes() (r []OTELAttribute) {
	if id := op.Spec.OperationID; id != "" {
		r = append(r, OTELAttribute{
			Key:   "OperationID",
			Value: id,
		})
	}
	if wh := op.WebhookInfo; wh != nil {
		r = append(r, OTELAttribute{
			Key:   "WebhookName",
			Value: wh.Name,
		})
	}
	return r
}

func (op Operation) PrettyOperationID() string {
	s := op.Spec
	if id := s.OperationID; id != "" {
		return id
	}
	var route string
	if info := op.WebhookInfo; info != nil {
		route = info.Name
	} else {
		route = s.Path.String()
	}
	return strings.ToUpper(s.HTTPMethod) + " " + route
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

func doTakePtr(t *Type) bool {
	return t.IsInterface() || t.DoPassByPointer()
}

func reqRespGoType(t *Type) string {
	if t.DoPassByPointer() {
		return "*" + t.Go()
	}
	return t.Go()
}

type Request struct {
	Type      *Type
	EmptyBody *Type
	Contents  map[ContentType]Media
	Spec      *openapi.RequestBody
}

// DoTakePtr returns true if type should be taken by pointer.
func (r *Request) DoTakePtr() bool {
	return doTakePtr(r.Type)
}

// GoType returns Go type of this response.
func (r *Request) GoType() string {
	return reqRespGoType(r.Type)
}

type Responses struct {
	Type       *Type
	Pattern    [5]*Response
	StatusCode map[int]*Response
	Default    *Response
}

// DoTakePtr returns true if type should be taken by pointer.
func (r *Responses) DoTakePtr() bool {
	return doTakePtr(r.Type)
}

// DoPass whether response type should be present in result tuple.
func (r *Responses) DoPass() bool {
	// In case of pattern responses or default response, the response type
	// has a StatusCode field, so it should be passed.
	if r.Default != nil || len(r.StatusCode) > 1 || r.HasPattern() {
		return true
	}
	t := r.Type
	// Do not pass response type if it is empty struct.
	if t.IsStruct() && len(t.Fields) == 0 {
		return false
	}
	return true
}

// GoType returns Go type of this response.
func (r *Responses) GoType() string {
	return reqRespGoType(r.Type)
}

// ResultTuple returns result tuple for this response.
func (r *Responses) ResultTuple(a, b string) string {
	if !r.DoPass() {
		return fmt.Sprintf("(%s error)", b)
	}
	typ := reqRespGoType(r.Type)
	if len(a)+len(b) > 0 {
		// Ensure that all result tuple elements are named
		// if any of them already is.
		if a == "" {
			a = "_"
		}
		if b == "" {
			b = "_"
		}
	}
	return fmt.Sprintf("(%s %s, %s error)", a, typ, b)
}

func (r *Responses) HasPattern() bool {
	for _, resp := range r.Pattern {
		if resp != nil {
			return true
		}
	}
	return false
}

type Response struct {
	NoContent *Type
	Contents  map[ContentType]Media
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

func (s Response) ResponseInfo(otel bool) []ResponseInfo {
	var result []ResponseInfo

	if noc := s.NoContent; noc != nil {
		result = append(result, ResponseInfo{
			Type:           noc,
			NoContent:      true,
			WithStatusCode: s.WithStatusCode,
			WithHeaders:    s.WithHeaders,
			OpenTelemetry:  otel,
			Headers:        s.Headers,
		})
	}
	for contentType, media := range s.Contents {
		result = append(result, ResponseInfo{
			Type:           media.Type,
			Encoding:       media.Encoding,
			ContentType:    contentType,
			WithStatusCode: s.WithStatusCode,
			WithHeaders:    s.WithHeaders,
			JSONStreaming:  media.JSONStreaming,
			OpenTelemetry:  otel,

			Headers: s.Headers,
		})
	}

	sortResponseInfos(result)
	return result
}
