package ir

import (
	"github.com/ogen-go/ogen/internal/oas"
)

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

// FirstRequiredIndex returns first required field index.
//
// Or -1 if there is no required fields.
func (j JSONFields) FirstRequiredIndex() int {
	for idx, f := range j {
		if typ := f.Type; typ.IsGeneric() && typ.GenericVariant.Optional ||
			typ.Is(
				KindStruct,
				KindEnum,
				KindPointer,
				KindSum,
				KindAlias,
			) && (typ.NilSemantic.Optional() || typ.NilSemantic.Invalid()) ||
			typ.IsArray() && typ.NilSemantic.Optional() {
			continue
		}
		return idx
	}
	return -1
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

// Format returns format name for handling json encoding or decoding.
//
// Mostly used for encoding or decoding of string formats, like `json.EncodeUUID`,
// where UUID is Format.
func (j JSON) Format() string {
	if j.t.Schema == nil {
		return ""
	}
	switch j.t.Schema.Format {
	case oas.FormatUUID:
		return "UUID"
	case oas.FormatDate:
		return "Date"
	case oas.FormatTime:
		return "Time"
	case oas.FormatDateTime:
		return "DateTime"
	case oas.FormatDuration:
		return "Duration"
	case oas.FormatIP, oas.FormatIPv4, oas.FormatIPv6:
		return "IP"
	case oas.FormatURI:
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
	if j.t.IsNumeric() {
		return "Number"
	}
	if j.t.Is(KindArray) {
		return "Array"
	}
	if j.t.Is(KindStruct) {
		return "Object"
	}
	switch j.t.Primitive {
	case Bool:
		return "Bool"
	case String, Time, Duration, UUID, IP, URL, ByteSlice:
		return "String"
	default:
		return ""
	}
}

// raw denotes whether Type can be encoded or decoded using simple
// json method, e.g. j.WriteString.
//
// Mostly true for primitives or enums.
func (j JSON) raw() bool {
	if !j.t.Is(KindPrimitive, KindEnum) {
		return false
	}

	if j.t.IsNumeric() {
		return true
	}
	switch j.t.Primitive {
	case Bool, String, ByteSlice:
		return true
	default:
		return false
	}
}

// Fn returns jx.Encoder or jx.Decoder method name.
//
// If blank, value cannot be encoded with single method call.
func (j JSON) Fn() string {
	if !j.raw() {
		return ""
	}
	switch j.t.Primitive {
	case String:
		return "Str"
	case ByteSlice:
		return "Base64"
	default:
		return capitalize(j.t.Primitive.String())
	}
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
