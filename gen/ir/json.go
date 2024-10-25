package ir

import (
	"slices"
	"strconv"
	"strings"

	"github.com/ogen-go/ogen/internal/bitset"
	"github.com/ogen-go/ogen/internal/naming"
	"github.com/ogen-go/ogen/internal/xslices"
	"github.com/ogen-go/ogen/jsonschema"
)

// JSON returns json encoding/decoding rules for t.
func (t *Type) JSON() JSON {
	return JSON{
		t: t,
	}
}

// JSON specifies json encoding and decoding for Type.
type JSON struct {
	t      *Type
	except []string
}

// AnyFields whether if type has any fields to encode.
func (j JSON) AnyFields() bool {
	for _, f := range j.t.Fields {
		if f.Inline != InlineNone {
			return true
		}

		t := f.Tag.JSON
		if t != "" && !slices.Contains(j.except, t) {
			return true
		}
	}
	return false
}

// Except return JSON with filter by given properties.
func (j JSON) Except(set ...string) JSON {
	return JSON{
		t:      j.t,
		except: set,
	}
}

type JSONFields []*Field

// FirstRequiredIndex returns first required field index.
//
// Or -1 if there is no required fields.
func (j JSONFields) FirstRequiredIndex() int {
	for idx, f := range j {
		if typ := f.Type; typ.IsGeneric() && typ.GenericVariant.Optional ||
			typ.Is(
				KindStruct,
				KindMap,
				KindEnum,
				KindPointer,
				KindSum,
				KindAlias,
			) && (typ.NilSemantic.Optional() || typ.NilSemantic.Invalid()) ||
			typ.IsArray() && typ.NilSemantic.Optional() ||
			typ.IsAny() {
			continue
		}
		return idx
	}
	return -1
}

// HasRequired whether object has required fields
func (j JSONFields) HasRequired() bool {
	return slices.ContainsFunc(j, func(f *Field) bool {
		return f.Spec != nil && f.Spec.Required
	})
}

// RequiredMask returns array of 64-bit bitmasks for required fields.
func (j JSONFields) RequiredMask() []uint8 {
	return bitset.Build(j, func(_ int, f *Field) bool {
		return f.Spec != nil && f.Spec.Required
	})
}

// Fields return all fields of Type that should be encoded via json.
func (j JSON) Fields() (fields JSONFields) {
	for _, f := range j.t.Fields {
		if t := f.Tag.JSON; t == "" || slices.Contains(j.except, t) {
			continue
		}
		fields = append(fields, f)
	}
	return fields
}

// AdditionalProps return field of Type that should be encoded as inlined map.
func (j JSON) AdditionalProps() *Field {
	f, _ := xslices.FindFunc(j.t.Fields, func(f *Field) bool {
		return f.Inline == InlineAdditional
	})
	return f
}

// PatternProps return field of Type that should be encoded as inlined map with pattern.
func (j JSON) PatternProps() (fields []*Field) {
	for _, f := range j.t.Fields {
		if f.Inline == InlinePattern {
			fields = append(fields, f)
		}
	}
	return fields
}

// SumProps return field of Type that should be encoded as inlined sum.
func (j JSON) SumProps() (fields []*Field) {
	for _, f := range j.t.Fields {
		if f.Inline == InlineSum {
			fields = append(fields, f)
		}
	}
	return fields
}

// Format returns format name for handling json encoding or decoding.
//
// Mostly used for encoding or decoding of string formats, like `json.EncodeUUID`,
// where UUID is Format.
func (j JSON) Format() string {
	s := j.t.Schema
	if s == nil {
		return ""
	}
	typePrefix := func(f string) string {
		switch s.Type {
		case jsonschema.String:
			return "String" + naming.Capitalize(f)
		default:
			return f
		}
	}
	switch f := s.Format; f {
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
	case "ip":
		return "IP"
	case "ipv4":
		return "IPv4"
	case "ipv6":
		return "IPv6"
	case "mac":
		return "MAC"
	case "uri":
		return "URI"
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64":
		if s.Type != jsonschema.String {
			return ""
		}
		return "String" + naming.Capitalize(f)
	case "unix", "unix-seconds":
		return typePrefix("UnixSeconds")
	case "unix-nano":
		return typePrefix("UnixNano")
	case "unix-micro":
		return typePrefix("UnixMicro")
	case "unix-milli":
		return typePrefix("UnixMilli")
	default:
		return ""
	}
}

// Type returns json value type that can represent Type.
//
// E.g. string primitive can be represented by StringValue which is commonly
// returned from `i.WhatIsNext()` method.
// Blank string is returned if there is no appropriate json type.
func (j JSON) Type() string {
	return jsonType(j.t)
}

func collectTypes(t *Type, types map[string]struct{}) {
	if !t.IsSum() {
		panic(unreachable(t))
	}
	for _, variant := range t.SumOf {
		typ := variant.JSON().Type()
		if typ == "" {
			collectTypes(variant, types)
			continue
		}
		types[typ] = struct{}{}
	}
}

// SumTypes returns jx.Type list for this sum type.
func (j JSON) SumTypes() string {
	types := map[string]struct{}{}
	collectTypes(j.t, types)

	sortedTypes := make([]string, 0, len(types))
	for k := range types {
		sortedTypes = append(sortedTypes, "jx."+k)
	}
	slices.Sort(sortedTypes)

	return strings.Join(sortedTypes, ",")
}

const arraySuffix = "Array"

func jsonType(t *Type) string {
	if t.IsNumeric() {
		if s := t.Schema; s != nil && s.Type == "string" {
			return "String"
		}
		return "Number"
	}
	if t.Is(KindArray) {
		return arraySuffix
	}
	if t.Is(KindStruct, KindMap) {
		return "Object"
	}
	if t.Is(KindGeneric) {
		return jsonType(t.GenericOf)
	}
	if t.Is(KindAlias) {
		return jsonType(t.AliasTo)
	}
	switch t.Primitive {
	case Bool:
		return "Bool"
	case Time:
		if s := t.Schema; s != nil && s.Type == "integer" {
			return "Number"
		}
		return "String"
	case String, Duration, UUID, MAC, IP, URL, ByteSlice:
		return "String"
	case Null:
		return "Null"
	default:
		return ""
	}
}

// raw denotes whether Type can be encoded or decoded using simple
// json method, e.g. j.WriteString.
//
// Mostly true for primitives or enums.
func (j JSON) raw() bool {
	if !j.t.Is(KindPrimitive, KindEnum, KindAny) {
		return false
	}

	if j.t.IsNumeric() {
		return true
	}
	switch j.t.Primitive {
	case Bool, String, ByteSlice:
		return true
	default:
		return j.t.Kind == KindAny
	}
}

func (j JSON) Decode() string {
	if j.t.IsAny() {
		// Copy to prevent referencing internal buffer.
		return "RawAppend(nil)"
	}
	// No arguments.
	return j.Fn() + "()"
}

// Fn returns jx.Encoder or jx.Decoder method name.
//
// If blank, value cannot be encoded with single method call.
func (j JSON) Fn() string {
	if !j.raw() {
		return ""
	}
	if j.t.IsAny() {
		return "Raw"
	}
	switch j.t.Primitive {
	case String:
		return "Str"
	case ByteSlice:
		return "Base64"
	case Uint,
		Uint8,
		Uint16,
		Uint32,
		Uint64:
		s := j.t.Primitive.String()
		return strings.ToUpper(s[:2]) + s[2:]
	default:
		return naming.Capitalize(j.t.Primitive.String())
	}
}

// IsBase64 whether field has base64 encoding.
func (j JSON) IsBase64() bool {
	return j.t.Primitive == ByteSlice
}

// TimeFormat returns time format for json encoding and decoding.
func (j JSON) TimeFormat() string {
	s := j.t.Schema
	if s == nil || s.XOgenTimeFormat == "" {
		return ""
	}
	return strconv.Quote(s.XOgenTimeFormat)
}

// Sum returns specification for parsing value as sum type.
func (j JSON) Sum() SumJSON {
	if j.t.SumSpec.Discriminator != "" {
		return SumJSON{
			Type: SumJSONDiscriminator,
		}
	}
	if j.t.SumSpec.TypeDiscriminator {
		return SumJSON{
			Type: SumJSONTypeDiscriminator,
		}
	}
	for _, s := range j.t.SumOf {
		if len(s.SumSpec.Unique) > 0 {
			return SumJSON{
				Type: SumJSONFields,
			}
		}
	}
	return SumJSON{
		Type: SumJSONPrimitive,
	}
}

type SumJSONType byte

const (
	SumJSONPrimitive SumJSONType = iota
	SumJSONFields
	SumJSONDiscriminator
	SumJSONTypeDiscriminator
)

// SumJSON specifies rules for parsing sum types in json.
type SumJSON struct {
	Type SumJSONType
}

func (s SumJSON) String() string {
	switch s.Type {
	case SumJSONFields:
		return "fields"
	case SumJSONPrimitive:
		return "primitive"
	case SumJSONDiscriminator:
		return "discriminator"
	case SumJSONTypeDiscriminator:
		return "type_discriminator"
	default:
		return "unknown"
	}
}

func (s SumJSON) Primitive() bool         { return s.Type == SumJSONPrimitive }
func (s SumJSON) Discriminator() bool     { return s.Type == SumJSONDiscriminator }
func (s SumJSON) TypeDiscriminator() bool { return s.Type == SumJSONTypeDiscriminator }
func (s SumJSON) Fields() bool            { return s.Type == SumJSONFields }
