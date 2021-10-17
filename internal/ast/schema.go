package ast

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/xerrors"

	"github.com/ogen-go/ogen/validate"
)

type SchemaKind = string

const (
	KindStruct    SchemaKind = "struct"
	KindAlias     SchemaKind = "alias"
	KindPrimitive SchemaKind = "primitive"
	KindArray     SchemaKind = "array"
	KindEnum      SchemaKind = "enum"
	KindPointer   SchemaKind = "pointer"
)

type Validators struct {
	String validate.String
	Int    validate.Int
	Array  validate.Array
}

// FormatCustom denotes that value type requires custom format handling
// for encoding and decoding.
const FormatCustom = "custom"

func (s Schema) FormatCustom() bool {
	return s.Format == FormatCustom
}

// JSONFields returns set of fields that should be encoded or decoded in json.
func (s Schema) JSONFields() []SchemaField {
	var fields []SchemaField
	for _, f := range s.Fields {
		if f.Tag == "-" {
			continue
		}
		fields = append(fields, f)
	}
	return fields
}

func (s Schema) IsStruct() bool {
	return s.Is(KindStruct)
}

func (s Schema) CanGeneric() bool {
	if !s.Is(KindPrimitive) {
		return false
	}
	if s.IsNumeric() {
		return true
	}
	switch s.Primitive {
	case "bool", "string", "time.Time", "time.Duration", "uuid.UUID":
		return true
	default:
		return false
	}
}

type Schema struct {
	Kind        SchemaKind
	Name        string
	Description string
	Doc         string
	Format      string

	AliasTo    *Schema
	PointerTo  *Schema
	Primitive  string
	Item       *Schema
	EnumValues []interface{}
	Fields     []SchemaField

	Implements map[*Interface]struct{}

	// Numeric validation.
	Validators Validators

	// String validation.
	// Pattern   string

	// Array validation.
	// UniqueItems bool

	// Struct validation.
	// MaxProperties *uint64
	// MinProperties *uint64

	Optional    bool
	Nullable    bool
	GenericType string
}

func (s Schema) canRawJSON() bool {
	if s.IsNumeric() {
		return true
	}
	switch s.Primitive {
	case "bool", "string", "time.Time", "time.Duration", "uuid.UUID":
		return true
	default:
		return false
	}
}

func (s Schema) JSONType() string {
	if s.IsNumeric() {
		return "NumberValue"
	}
	switch s.Primitive {
	case "bool":
		return "BoolValue"
	case "string", "time.Time", "time.Duration", "uuid.UUID":
		return "StringValue"
	default:
		return ""
	}
}

// JSONFormat returns format name for handling json encoding or decoding.
//
// Mostly used for encoding or decoding of generics, like "json.WriteUUID",
// where UUID is JSONFormat.
func (s Schema) JSONFormat() string {
	switch s.Format {
	case "uuid":
		return "UUID"
	case "date":
		return "Date"
	case "time":
		return "Time"
	case "date-time":
		return "DateTime"
	case "duration":
		return "Duration"
	default:
		return ""
	}
}

func capitalize(s string) string {
	var v []rune
	for i, c := range s {
		if i == 0 {
			c = unicode.ToUpper(c)
		}
		v = append(v, c)
	}
	return string(v)
}

func (s Schema) jsonFn() string {
	if !s.canRawJSON() {
		return ""
	}
	return capitalize(s.Primitive)
}

// JSONWrite returns function name from json package that writes value.
func (s Schema) JSONWrite() string {
	if s.jsonFn() == "" {
		return ""
	}
	return "Write" + s.jsonFn()
}

// JSONRead returns function name from json package that reads value.
func (s Schema) JSONRead() string {
	if s.jsonFn() == "" {
		return ""
	}
	return "Read" + s.jsonFn()
}

// Generic returns true if value is generic type, e.g. nullable or optional.
func (s Schema) Generic() bool {
	if s.Optional {
		return true
	}
	if s.Nullable {
		return true
	}
	return false
}

// GenericKind returns name of generic kind.
// Used in generic type names.
func (s Schema) GenericKind() string {
	var b strings.Builder
	if s.Optional {
		b.WriteString("Optional")
	}
	if s.Nullable {
		b.WriteString("Nil")
	}
	return b.String()
}

func (s *Schema) IsArray() bool {
	return s.Is(KindArray)
}

func (s *Schema) IsInteger() bool {
	switch s.Primitive {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return true
	default:
		return false
	}
}

func (s *Schema) IsFloat() bool {
	switch s.Primitive {
	case "float32", "float64":
		return true
	default:
		return false
	}
}

func (s *Schema) IsNumeric() bool { return s.IsInteger() || s.IsFloat() }

func (s *Schema) NeedValidation() bool {
	return s.needValidation(map[*Schema]struct{}{})
}

func (s *Schema) needValidation(visited map[*Schema]struct{}) (result bool) {
	if s == nil {
		return false
	}

	if _, ok := visited[s]; ok {
		return false
	}

	visited[s] = struct{}{}

	switch s.Kind {
	case KindPrimitive:
		if s.Generic() {
			// TODO(ernado): fix
			return false
		}
		if s.IsNumeric() && s.Validators.Int.Set() {
			return true
		}
		if s.Validators.String.Set() {
			return true
		}
		return false
	case KindEnum:
		return true
	case KindAlias:
		return s.AliasTo.needValidation(visited)
	case KindPointer:
		return s.PointerTo.needValidation(visited)
	case KindArray:
		if s.Validators.Array.Set() {
			return true
		}
		// Prevent infinite recursion.
		if s.Item == s {
			return false
		}
		return s.Item.needValidation(visited)
	case KindStruct:
		for _, f := range s.Fields {
			if f.Type.needValidation(visited) {
				return true
			}
		}
		return false
	default:
		panic("unreachable")
	}
}

type SchemaField struct {
	Name string
	Type *Schema
	Tag  string
}

func afterDot(v string) string {
	idx := strings.Index(v, ".")
	if idx > 0 {
		return v[idx+1:]
	}
	return v
}

func (s Schema) EncodeFn() string {
	if s.IsArray() && s.Item.EncodeFn() != "" {
		return s.Item.EncodeFn() + "Array"
	}
	switch s.Primitive {
	case "interface{}":
		return "Interface"
	case "int", "int64", "int32", "string", "bool", "float32", "float64":
		return capitalize(s.Primitive)
	case "uuid.UUID", "time.Time":
		return afterDot(s.Primitive)
	default:
		return ""
	}
}

func (s Schema) ToString() string {
	if s.EncodeFn() == "" {
		return ""
	}
	return s.EncodeFn() + "ToString"
}

func (s Schema) FromString() string {
	if s.EncodeFn() == "" {
		return ""
	}
	return "To" + s.EncodeFn()
}

func (s Schema) Type() string {
	switch s.Kind {
	case KindStruct:
		return s.Name
	case KindAlias:
		return s.Name
	case KindPrimitive:
		if s.Generic() {
			return s.GenericType
		}
		return s.Primitive
	case KindArray:
		return "[]" + s.Item.Type()
	case KindEnum:
		return s.Name
	case KindPointer:
		return "*" + s.PointerTo.Type()
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
	if s.Is(KindPrimitive, KindArray, KindPointer) {
		panic("unreachable")
	}

	if s.Implements == nil {
		s.Implements = map[*Interface]struct{}{}
	}
	if iface.Implementations == nil {
		iface.Implementations = map[*Schema]struct{}{}
	}

	iface.Implementations[s] = struct{}{}
	s.Implements[iface] = struct{}{}
}

func (s *Schema) Unimplement(iface *Interface) {
	delete(iface.Implementations, s)
	delete(s.Implements, iface)
}

func (s *Schema) Methods() []string {
	ms := make(map[string]struct{})
	for iface := range s.Implements {
		for m := range iface.Methods {
			ms[m] = struct{}{}
		}
	}

	var result []string
	for m := range ms {
		result = append(result, m)
	}
	sort.Strings(result)
	return result
}

func Struct(name string) *Schema {
	return &Schema{
		Kind: KindStruct,
		Name: name,
	}
}

func Primitive(typ string) *Schema {
	return &Schema{
		Kind:      KindPrimitive,
		Primitive: typ,
	}
}

func Alias(name string, typ *Schema) *Schema {
	return &Schema{
		Kind:    KindAlias,
		Name:    name,
		AliasTo: typ,
	}
}

func Pointer(to *Schema) *Schema {
	return &Schema{
		Kind:      KindPointer,
		PointerTo: to,
	}
}

func Array(item *Schema) *Schema {
	return &Schema{
		Kind: KindArray,
		Item: item,
	}
}

func Enum(name, typ string, rawValues []json.RawMessage) (*Schema, error) {
	var (
		values []interface{}
		uniq   = map[interface{}]struct{}{}
	)
	for _, raw := range rawValues {
		val, err := parseJsonValue(typ, raw)
		if err != nil {
			if xerrors.Is(err, errNullValue) {
				continue
			}
			return nil, xerrors.Errorf("parse value '%s': %w", raw, err)
		}

		if _, found := uniq[val]; found {
			return nil, xerrors.Errorf("duplicate enum value: '%v'", val)
		}

		uniq[val] = struct{}{}
		values = append(values, val)
	}

	return &Schema{
		Kind:       KindEnum,
		Name:       name,
		Primitive:  typ,
		EnumValues: values,
	}, nil
}

func Iface(name string) *Interface {
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
