package ast

import "fmt"

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

	AliasTo   *Schema
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
		return s.Name
	case KindPrimitive:
		return s.Primitive
	case KindArray:
		return "[]" + s.Item.Type()
	default:
		panic(fmt.Errorf("unexpected SchemaKind: %s", s.Kind))
	}
}

func (s Schema) Is(vs ...SchemaKind) bool {
	for _, v := range vs {
		if s.Kind == v {
			return true
		}
	}

	return false
}

func (s *Schema) Implement(iface *Interface) {
	if s.Is(KindPrimitive, KindArray) {
		panic("unreachable")
	}

	if s.Implements == nil {
		s.Implements = map[string]struct{}{}
	}
	if iface.Implementations == nil {
		iface.Implementations = map[*Schema]struct{}{}
	}

	iface.Implementations[s] = struct{}{}
	for method := range iface.Methods {
		s.Implements[method] = struct{}{}
	}
}

func (s *Schema) Unimplement(iface *Interface) {
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

func CreateSchemaStruct(name string) *Schema {
	return &Schema{
		Kind: KindStruct,
		Name: name,
	}
}

func CreateSchemaPrimitive(typ string) *Schema {
	return &Schema{
		Kind:      KindPrimitive,
		Primitive: typ,
	}
}

func CreateSchemaAlias(name string, typ *Schema) *Schema {
	return &Schema{
		Kind:    KindAlias,
		Name:    name,
		AliasTo: typ,
	}
}

func CreateSchemaArray(item *Schema) *Schema {
	return &Schema{
		Kind: KindArray,
		Item: item,
	}
}

func CreateIface(name string) *Interface {
	return &Interface{
		Name:            name,
		Methods:         map[string]struct{}{},
		Implementations: map[*Schema]struct{}{},
	}
}

func CreateRequestBody() *RequestBody {
	return &RequestBody{
		Contents: map[string]*Schema{},
	}
}

func CreateResponse() *Response {
	return &Response{
		Contents: map[string]*Schema{},
	}
}
