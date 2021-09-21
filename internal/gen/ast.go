package gen

import "strings"

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
	Path       string
	HTTPMethod string
	Parameters map[ParameterLocation][]Parameter

	RequestType string
	RequestBody *RequestBody

	ResponseType    string
	Responses       map[int]*Response
	ResponseDefault *Response
}

type Parameter struct {
	Name       string
	SourceName string
	Type       string
	In         ParameterLocation

	// In - [Possible style values]
	//   "path"   - "simple", "label", "matrix".
	//   "query"  - "form", "spaceDelimited", "pipeDelimited", "deepObject".
	//   "header" - "simple".
	//   "cookie" - "form".
	// Style string

	// Explode bool

	Required bool
}

type Schema struct {
	Name        string
	Description string

	Simple string
	Fields []SchemaField

	Implements map[string]struct{}
}

func (s Schema) typeName() string {
	if len(s.Fields) > 0 {
		return s.Name
	}
	return s.Simple
}

func (g *Generator) createSchema(name string) *Schema {
	s := &Schema{
		Name:       name,
		Implements: map[string]struct{}{},
	}
	g.schemas[name] = s
	return s
}

func (g *Generator) createSchemaSimple(name, typ string) *Schema {
	s := g.createSchema(name)
	s.Simple = typ
	return s
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

func (g *Generator) createResponse(name string) *Response {
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
				Default: true,
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
