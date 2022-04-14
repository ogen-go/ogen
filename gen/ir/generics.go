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
	switch t.Kind {
	case KindStruct:
		if len(t.Fields) == 0 {
			return false
		}
	case KindAlias:
		return t.AliasTo.CanGeneric()
	case KindSum:
		if len(t.SumOf) == 0 {
			return false
		}
	}
	return t.Is(KindPrimitive, KindEnum, KindStruct, KindMap, KindSum)
}
