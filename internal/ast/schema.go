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
	// KindPointer simulates optionals via go pointers.
	// Deprecated. Use KindGeneric.
	KindPointer SchemaKind = "pointer"
	KindGeneric SchemaKind = "generic"
)

type Validators struct {
	String validate.String
	Int    validate.Int
	Array  validate.Array
}

func (s Schema) FormatCustom() bool {
	switch s.Primitive {
	case "time.Time":
		return true
	default:
		return false
	}
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

func (s Schema) IsGeneric() bool {
	return s.Is(KindGeneric)
}

func (s Schema) CanGeneric() bool {
	if s.Primitive == "[]byte" || s.Type() == "struct{}" {
		return false
	}
	return s.Is(KindPrimitive, KindEnum, KindStruct)
}

// ArrayVariant specifies nil value semantics of slice.
type ArrayVariant string

// Possible Array nil semantics.
const (
	ArrayRequired ArrayVariant = "required" // nil is invalid
	ArrayOptional ArrayVariant = "optional" // nil is "no value"
	ArrayNullable ArrayVariant = "nullable" // nil is null
)

type Schema struct {
	Kind        SchemaKind
	Name        string
	Description string
	Doc         string
	Format      string

	GenericOf      *Schema
	GenericVariant GenericVariant

	AliasTo   *Schema
	PointerTo *Schema
	Primitive string

	Item         *Schema
	ArrayVariant ArrayVariant

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
}

func (s Schema) canRawJSON() bool {
	if s.IsNumeric() {
		return true
	}
	switch s.Primitive {
	case "bool", "string":
		return true
	default:
		return false
	}
}

func (s Schema) JSONType() string {
	if s.IsNumeric() {
		return "NumberValue"
	}
	if s.IsArray() {
		return "ArrayValue"
	}
	if s.IsStruct() {
		return "ObjectValue"
	}
	switch s.Primitive {
	case "bool":
		return "BoolValue"
	case "string", "time.Time", "time.Duration", "uuid.UUID", "net.IP", "url.URL":
		return "StringValue"
	default:
		return ""
	}
}

// JSONHelper returns format name for handling json encoding or decoding.
//
// Mostly used for encoding or decoding of generics, like "json.WriteUUID",
// where UUID is JSONHelper.
func (s Schema) JSONHelper() string {
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
	case "ip", "ipv4", "ipv6":
		return "IP"
	case "uri":
		return "URI"
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

func (s *Schema) IsArray() bool {
	return s.Is(KindArray)
}

func (s *Schema) IsEnum() bool {
	return s.Is(KindEnum)
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
	case KindGeneric:
		return s.GenericOf.needValidation(visited)
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
		return s.Primitive
	case KindGeneric:
		return s.Name
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

// Pointer makes new pointer type.
//
// Deprecated, use generics.
func Pointer(to *Schema) *Schema {
	return &Schema{
		Kind:      KindPointer,
		PointerTo: to,
	}
}

type GenericVariant struct {
	Nullable bool
	Optional bool
}

func (v GenericVariant) OnlyOptional() bool {
	return v.Optional && !v.Nullable
}

func (v GenericVariant) OnlyNullable() bool {
	return v.Nullable && !v.Optional
}

func (v GenericVariant) Name() string {
	var b strings.Builder
	if v.Optional {
		b.WriteString("Opt")
	}
	if v.Nullable {
		b.WriteString("Nil")
	}
	return b.String()
}

func (v GenericVariant) Any() bool {
	return v.Nullable || v.Optional
}

func Generic(name string, of *Schema, v GenericVariant) *Schema {
	name = v.Name() + name
	if of.IsArray() {
		name = name + "Array"
	}
	return &Schema{
		Name:           name,
		Kind:           KindGeneric,
		GenericOf:      of,
		GenericVariant: v,
	}
}

func Array(item *Schema) *Schema {
	return &Schema{
		Kind:         KindArray,
		Item:         item,
		ArrayVariant: ArrayRequired,
	}
}

func Enum(name, typ string, rawValues []json.RawMessage) (*Schema, error) {
	var (
		values []interface{}
		uniq   = map[interface{}]struct{}{}
	)
	for _, raw := range rawValues {
		val, err := parseJSONValue(typ, raw)
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
