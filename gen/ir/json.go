package ir

import "github.com/ogen-go/ogen/internal/capitalize"

// JSON returns json encoding/decoding rules for t.
func (t *Type) JSON() JSON {
	return JSON{
		t: t,
	}
}

// JSON specifies json encoding and decoding for Type.
type JSON struct {
	t *Type
}

type JSONFields []*Field

// NotEmpty whether field slice is not empty.
func (j JSONFields) NotEmpty() bool {
	return len(j) != 0
}

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
	for _, f := range j {
		if f.Spec != nil && f.Spec.Required {
			return true
		}
	}
	return false
}

// RequiredMask returns array of 64-bit bitmasks for required fields.
func (j JSONFields) RequiredMask() (r []uint8) {
	i := 0
	r = append(r, 0)
	for _, f := range j {
		maskIdx := i / 8
		if len(r) <= maskIdx {
			r = append(r, 0)
		}
		bitIdx := i % 8

		set := uint8(0)
		if f.Spec != nil && f.Spec.Required {
			set = 1
		}
		r[maskIdx] |= set << uint8(bitIdx)
		i++
	}
	return r
}

// Fields return all fields of Type that should be encoded via json.
func (j JSON) Fields() (fields JSONFields) {
	for _, f := range j.t.Fields {
		if f.Tag.JSON == "" {
			continue
		}
		fields = append(fields, f)
	}
	return fields
}

// AdditionalProps return field of Type that should be encoded as inlined map.
func (j JSON) AdditionalProps() (field *Field) {
	for _, f := range j.t.Fields {
		if f.Inline == InlineAdditional {
			return f
		}
	}
	return nil
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

// Format returns format name for handling json encoding or decoding.
//
// Mostly used for encoding or decoding of string formats, like `json.EncodeUUID`,
// where UUID is Format.
func (j JSON) Format() string {
	if j.t.Schema == nil {
		return ""
	}
	switch j.t.Schema.Format {
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

// Type returns json value type that can represent Type.
//
// E.g. string primitive can be represented by StringValue which is commonly
// returned from `i.WhatIsNext()` method.
// Blank string is returned if there is no appropriate json type.
func (j JSON) Type() string {
	return jsonType(j.t)
}

func jsonType(t *Type) string {
	if t.IsNumeric() {
		return "Number"
	}
	if t.Is(KindArray) {
		return "Array"
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
	case String, Time, Duration, UUID, IP, URL, ByteSlice:
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
	default:
		return capitalize.Capitalize(j.t.Primitive.String())
	}
}

// IsBase64 whether field has base64 encoding.
func (j JSON) IsBase64() bool {
	return j.t.Primitive == ByteSlice
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
