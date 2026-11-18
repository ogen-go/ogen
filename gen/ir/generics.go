package ir

import "strings"

type GenericVariant struct {
	Nullable bool
	Optional bool
}

func (v GenericVariant) NullableOptional() bool {
	return v.Optional && v.Nullable
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

// CanGeneric reports whether Type can be boxed to KindGeneric.
func (t Type) CanGeneric() bool {
	return t.canGeneric(false)
}

// CanGenericNullable reports whether Type can be boxed to KindGeneric if the value can be null.
//
// Unlike CanGeneric, it accepts structs without fields. A nil pointer to such a struct
// cannot be told apart from an empty object, so null needs the explicit flag of a generic.
func (t Type) CanGenericNullable() bool {
	return t.canGeneric(true)
}

func (t Type) canGeneric(emptyStruct bool) bool {
	switch t.Kind {
	case KindStruct:
		if len(t.Fields) == 0 && !emptyStruct {
			return false
		}
	case KindAlias:
		return t.AliasTo.canGeneric(emptyStruct)
	case KindSum:
		if len(t.SumOf) == 0 {
			return false
		}
	}
	return t.Is(KindPrimitive, KindEnum, KindStruct, KindMap, KindSum)
}
