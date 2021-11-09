package ir

import (
	"fmt"
	"strings"
	"unicode"
)

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

func afterDot(v string) string {
	idx := strings.Index(v, ".")
	if idx > 0 {
		return v[idx+1:]
	}
	return v
}

func (t *Type) EncodeFn() string {
	switch t.Kind {
	case KindArray:
		if t.Array.Item.EncodeFn() != "" {
			return t.Array.Item.EncodeFn() + "Array"
		}
	case KindPrimitive:
		switch t.Primitive.Type {
		case Int, Int64, Int32, String, Bool, Float32, Float64:
			return capitalize(t.Primitive.Type.String())
		case UUID, Time, IP, Duration, URL:
			return afterDot(t.Primitive.Type.String())
		}
	case KindEnum:
		switch t.Enum.Type {
		case Int, Int64, Int32, String, Bool, Float32, Float64:
			return capitalize(t.Enum.Type.String())
		case UUID, Time, IP, Duration, URL:
			return afterDot(t.Enum.Type.String())
		}
	}
	return ""
}

func (t Type) ToString() string {
	if t.EncodeFn() == "" {
		return ""
	}
	return t.EncodeFn() + "ToString"
}

func (t Type) FromString() string {
	if t.EncodeFn() == "" {
		return ""
	}
	return "To" + t.EncodeFn()
}

func (t *Type) IsInteger() bool {
	if t.Kind == KindPrimitive {
		switch t.Primitive.Type {
		case Int, Int8, Int16, Int32, Int64,
			Uint, Uint8, Uint16, Uint32, Uint64:
			return true
		}
	}
	return false
}

func (t *Type) IsFloat() bool {
	if t.Kind == KindPrimitive {
		switch t.Primitive.Type {
		case Float32, Float64:
			return true
		}
	}
	return false
}

func (t *Type) IsPrimitive() bool { return t.Is(KindPrimitive) }
func (t *Type) IsArray() bool     { return t.Is(KindArray) }
func (t *Type) IsStruct() bool    { return t.Is(KindStruct) }
func (t *Type) IsPointer() bool   { return t.Is(KindPointer) }
func (t *Type) IsEnum() bool      { return t.Is(KindEnum) }
func (t *Type) IsGeneric() bool   { return t.Is(KindGeneric) }
func (t *Type) IsAlias() bool     { return t.Is(KindAlias) }
func (t *Type) IsInterface() bool { return t.Is(KindInterface) }
func (t *Type) IsSum() bool       { return t.Is(KindSum) }
func (t *Type) IsStream() bool    { return t.Is(KindStream) }
func (t *Type) IsNumeric() bool   { return t.IsInteger() || t.IsFloat() }

func (t *Type) MustField(name string) *Field {
	if !t.Is(KindStruct) {
		panic("unreachable")
	}

	for _, f := range t.Struct.Fields {
		if f.Name == name {
			return f
		}
	}

	panic(fmt.Sprintf("field with name %q not found", name))
}

func (t *Type) MustName() string {
	switch t.Kind {
	case KindStruct:
		return t.Struct.Name
	case KindAlias:
		return t.Alias.Name
	case KindEnum:
		return t.Enum.Name
	case KindGeneric:
		return t.Generic.Name
	case KindSum:
		return t.Sum.Name
	case KindStream:
		return t.Stream.Name
	case KindInterface:
		return t.Interface.Name
	default:
		// fmt.Printf("WARN: Cannot get name from %q\n", t.Kind)
		return t.ghostName
	}
}

func (t *Type) MustSetName(name string) {
	switch t.Kind {
	case KindStruct:
		t.Struct.Name = name
	case KindAlias:
		t.Alias.Name = name
	case KindEnum:
		t.Enum.Name = name
	case KindGeneric:
		t.Generic.Name = name
	case KindSum:
		t.Sum.Name = name
	case KindStream:
		t.Stream.Name = name
	case KindInterface:
		t.Interface.Name = name
	default:
		// fmt.Printf("WARN: Cannot set name to %q\n", t.Kind)
		t.ghostName = name
	}
}
