package gen

type TemplateConfig struct {
	Package    string
	Methods    []*Method
	Schemas    map[string]*Schema
	Interfaces map[string]*Interface
}

type ParameterLocation string

const (
	LocationQuery  ParameterLocation = "Query"
	LocationHeader ParameterLocation = "Header"
	LocationPath   ParameterLocation = "Path"
	LocationCookie ParameterLocation = "Cookie"
)

type Method struct {
	Name       string
	Path       string
	HTTPMethod string
	Parameters map[ParameterLocation][]Parameter

	RequestType string

	RequestBody *RequestBody
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
	Name    string
	Methods map[string]struct{}
	Schemas map[*Schema]struct{}
}

type RequestBody struct {
	Contents map[string]*Schema
	Required bool
}
