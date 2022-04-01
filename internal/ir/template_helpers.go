package ir

import (
	"fmt"
	"strings"

	"github.com/ogen-go/ogen/internal/capitalize"
)

func afterDot(v string) string {
	idx := strings.Index(v, ".")
	if idx > 0 {
		return v[idx+1:]
	}
	return v
}

func (t *Type) EncodeFn() string {
	if t.Is(KindArray) && t.Item.EncodeFn() != "" {
		return t.Item.EncodeFn() + "Array"
	}
	switch t.Primitive {
	case ByteSlice:
		return "Bytes"
	case Int, Int64, Int32, String, Bool, Float32, Float64:
		return capitalize.Capitalize(t.Primitive.String())
	case UUID, Time, IP, Duration, URL:
		return afterDot(t.Primitive.String())
	default:
		return ""
	}
}

func (t Type) ToString() string {
	encodeFn := t.EncodeFn()
	if encodeFn == "" {
		panic(fmt.Sprintf("unexpected %+v", t))
	}
	return encodeFn + "ToString"
}

func (t Type) FromString() string {
	encodeFn := t.EncodeFn()
	if encodeFn == "" {
		panic(fmt.Sprintf("unexpected %+v", t))
	}
	return "To" + encodeFn
}

func (t *Type) IsInteger() bool {
	switch t.Primitive {
	case Int, Int8, Int16, Int32, Int64,
		Uint, Uint8, Uint16, Uint32, Uint64:
		return true
	default:
		return false
	}
}

func (t *Type) IsFloat() bool {
	switch t.Primitive {
	case Float32, Float64:
		return true
	default:
		return false
	}
}

func (t *Type) IsArray() bool     { return t.Is(KindArray) }
func (t *Type) IsMap() bool       { return t.Is(KindMap) }
func (t *Type) IsPrimitive() bool { return t.Is(KindPrimitive) }
func (t *Type) IsStruct() bool    { return t.Is(KindStruct) }
func (t *Type) IsPointer() bool   { return t.Is(KindPointer) }
func (t *Type) IsEnum() bool      { return t.Is(KindEnum) }
func (t *Type) IsGeneric() bool   { return t.Is(KindGeneric) }
func (t *Type) IsAlias() bool     { return t.Is(KindAlias) }
func (t *Type) IsInterface() bool { return t.Is(KindInterface) }
func (t *Type) IsSum() bool       { return t.Is(KindSum) }
func (t *Type) IsAny() bool       { return t.Is(KindAny) }
func (t *Type) IsStream() bool    { return t.Is(KindStream) }
func (t *Type) IsNumeric() bool   { return t.IsInteger() || t.IsFloat() }

func (t *Type) ReceiverType() string {
	if t.needsPointerReceiverType() {
		return "*" + t.Name
	}
	return t.Name
}

func (t *Type) needsPointerReceiverType() bool {
	switch t.Kind {
	case KindPointer, KindArray, KindMap:
		return false
	case KindAlias:
		return t.AliasTo.needsPointerReceiverType()
	default:
		return true
	}
}

func (t *Type) MustField(name string) *Field {
	if !t.Is(KindStruct) {
		panic(unreachable(t))
	}

	for _, f := range t.Fields {
		if f.Name == name {
			return f
		}
	}

	panic(fmt.Sprintf("field with name %q not found", name))
}

func (t *Type) SetFieldType(name string, newT *Type) {
	if !t.Is(KindStruct) {
		panic(unreachable(t))
	}

	for _, f := range t.Fields {
		if f.Name == name {
			f.Type = newT
			return
		}
	}

	panic(fmt.Sprintf("field with name %q not found", name))
}
