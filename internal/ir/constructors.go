package ir

import "github.com/ogen-go/ogen/internal/oas"

func Primitive(typ PrimitiveType, schema *oas.Schema) *Type {
	return &Type{
		Kind:      KindPrimitive,
		Primitive: typ,
		Schema:    schema,
	}
}

func Array(item *Type, schema *oas.Schema) *Type {
	return &Type{
		Kind:   KindArray,
		Item:   item,
		Schema: schema,
	}
}

func Alias(name string, to *Type) *Type {
	return &Type{
		Kind:    KindAlias,
		Name:    name,
		AliasTo: to,
	}
}

func Interface(name string) *Type {
	return &Type{
		Name:             name,
		Kind:             KindInterface,
		InterfaceMethods: map[string]struct{}{},
		Implementations:  map[*Type]struct{}{},
	}
}

func Pointer(typ *Type, sem NilSemantic) *Type {
	return &Type{
		Kind:        KindPointer,
		PointerTo:   typ,
		NilSemantic: sem,
	}
}
