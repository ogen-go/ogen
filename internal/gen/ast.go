package gen

import (
	"fmt"
	"strings"
)

type ParameterLocation string

const (
	LocationQuery  ParameterLocation = "Query"
	LocationHeader ParameterLocation = "Header"
	LocationPath   ParameterLocation = "Path"
	LocationCookie ParameterLocation = "Cookie"
)

func (p ParameterLocation) Lower() string { return strings.ToLower(string(p)) }

type Method struct {
	Name       string
	PathParts  []PathPart
	HTTPMethod string
	Parameters []*Parameter

	RequestType string
	RequestBody *RequestBody

	ResponseType    string
	Responses       map[int]*Response
	ResponseDefault *Response
}

func (m *Method) PathParams() []*Parameter   { return m.getParams(LocationPath) }
func (m *Method) QueryParams() []*Parameter  { return m.getParams(LocationQuery) }
func (m *Method) CookieParams() []*Parameter { return m.getParams(LocationCookie) }
func (m *Method) HeaderParams() []*Parameter { return m.getParams(LocationHeader) }

func (m *Method) getParams(locatedIn ParameterLocation) []*Parameter {
	var params []*Parameter
	for _, p := range m.Parameters {
		if p.In == locatedIn {
			params = append(params, p)
		}
	}
	return params
}

type PathPart struct {
	Raw   string
	Param *Parameter
}

func (m *Method) Path() string {
	var path string
	for _, part := range m.PathParts {
		if part.Raw != "" {
			path += "/" + part.Raw
			continue
		}

		path += "/{" + part.Param.SourceName + "}"
	}
	return path
}

type Parameter struct {
	Name       string
	SourceName string
	Schema     *Schema
	In         ParameterLocation
	Style      string
	Explode    bool

	Required bool
}

type SchemaKind = string

const (
	KindStruct    SchemaKind = "struct"
	KindAlias     SchemaKind = "alias"
	KindPrimitive SchemaKind = "primitive"
	KindArray     SchemaKind = "array"
)

type Schema struct {
	Kind        SchemaKind
	Name        string
	Description string

	AliasTo   string
	Primitive string
	Item      *Schema
	Fields    []SchemaField

	Implements map[string]struct{}
}

func (s Schema) Type() string {
	switch s.Kind {
	case KindStruct:
		return s.Name
	case KindAlias:
		return s.AliasTo
	case KindPrimitive:
		return s.Primitive
	case KindArray:
		return "[]" + s.Item.Type()
	default:
		panic(fmt.Errorf("unexpected SchemaKind: %s", s.Kind))
	}
}

func (g *Generator) createSchemaStruct(name string) *Schema {
	return &Schema{
		Kind:       KindStruct,
		Name:       name,
		Implements: map[string]struct{}{},
	}
}

func (g *Generator) createSchemaPrimitive(name, typ string) *Schema {
	return &Schema{
		Kind:       KindPrimitive,
		Name:       name,
		Primitive:  typ,
		Implements: map[string]struct{}{},
	}
}

func (g *Generator) createSchemaAlias(name, typ string) *Schema {
	return &Schema{
		Kind:       KindAlias,
		Name:       name,
		AliasTo:    typ,
		Implements: map[string]struct{}{},
	}
}

func (g *Generator) createSchemaArray(name string, item *Schema) *Schema {
	return &Schema{
		Kind:       KindArray,
		Name:       name,
		Item:       item,
		Implements: map[string]struct{}{},
	}
}

func (s *Schema) implement(iface *Interface) {
	iface.Implementations[s] = struct{}{}
	for method := range iface.Methods {
		s.Implements[method] = struct{}{}
	}
}

func (s *Schema) unimplement(iface *Interface) {
	delete(iface.Implementations, s)
	for method := range iface.Methods {
		delete(s.Implements, method)
	}
}

func (s Schema) EqualFields(another Schema) bool {
	if len(s.Fields) != len(another.Fields) {
		return false
	}

	for i := 0; i < len(s.Fields); i++ {
		l, r := s.Fields[i], another.Fields[i]
		if l.Name != r.Name || l.Type != r.Type || l.Tag != r.Tag {
			return false
		}
	}

	return true
}

type SchemaField struct {
	Name string
	Tag  string
	Type string
}

type Interface struct {
	Name            string
	Methods         map[string]struct{}
	Implementations map[*Schema]struct{}
}

func (g *Generator) createIface(name string) *Interface {
	iface := &Interface{
		Name:            name,
		Methods:         map[string]struct{}{},
		Implementations: map[*Schema]struct{}{},
	}
	g.interfaces[name] = iface
	return iface
}

func (i *Interface) addMethod(method string) {
	i.Methods[method] = struct{}{}
	for schema := range i.Implementations {
		schema.Implements[method] = struct{}{}
	}
}

type RequestBody struct {
	Contents map[string]*Schema
	Required bool
}

func (g *Generator) createRequestBody() *RequestBody {
	return &RequestBody{
		Contents: map[string]*Schema{},
	}
}

type Response struct {
	NoContent *Schema
	Contents  map[string]*Schema
}

func (g *Generator) createResponse() *Response {
	return &Response{
		Contents: map[string]*Schema{},
	}
}

func (r *Response) implement(iface *Interface) {
	if s := r.NoContent; s != nil {
		s.implement(iface)
	}

	for _, schema := range r.Contents {
		schema.implement(iface)
	}
}

func (r *Response) unimplement(iface *Interface) {
	if s := r.NoContent; s != nil {
		s.unimplement(iface)
	}

	for _, schema := range r.Contents {
		schema.unimplement(iface)
	}
}

type ResponseInfo struct {
	StatusCode  int
	ContentType string
	NoContent   bool
	Default     bool
}

func (m *Method) ListResponseSchemas() map[*Schema]ResponseInfo {
	schemas := make(map[*Schema]ResponseInfo)
	for statusCode, resp := range m.Responses {
		if resp.NoContent != nil {
			schemas[resp.NoContent] = ResponseInfo{
				StatusCode: statusCode,
				NoContent:  true,
			}
			continue
		}
		for contentType, schema := range resp.Contents {
			schemas[schema] = ResponseInfo{
				StatusCode:  statusCode,
				ContentType: contentType,
			}
		}
	}

	if def := m.ResponseDefault; def != nil {
		if noc := def.NoContent; noc != nil {
			schemas[noc] = ResponseInfo{
				Default:   true,
				NoContent: true,
			}
		}
		for contentType, schema := range def.Contents {
			schemas[schema] = ResponseInfo{
				Default:     true,
				ContentType: contentType,
			}
		}
	}
	return schemas
}
