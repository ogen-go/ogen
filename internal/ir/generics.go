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

func Generic(name string, of *Type, v GenericVariant) *Type {
	name = v.Name() + name
	if of.Is(KindArray) {
		name = name + "Array"
	}
	return &Type{
		Kind: KindGeneric,
		Generic: &TypeGeneric{
			Name:    name,
			Of:      of,
			Variant: v,
		},
	}
}

// CanGeneric reports whether Type can be boxed to KindGeneric.
func (t Type) CanGeneric() bool {
	switch t.Kind {
	case KindStruct:
		if len(t.Struct.Fields) == 0 {
			return false
		}
	case KindArray:
		if t.Array.Item.Kind == KindPrimitive {
			if t.Array.Item.Primitive.Type == Byte {
				return false
			}
		}
	case KindAlias:
		return t.Alias.To.CanGeneric()
	}
	return t.Is(KindPrimitive, KindEnum, KindStruct)
}
