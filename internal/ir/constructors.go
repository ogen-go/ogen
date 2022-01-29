package ir

import "github.com/ogen-go/ogen/internal/oas"

func Primitive(typ PrimitiveType, schema *oas.Schema) *Type {
	return &Type{
		Kind:      KindPrimitive,
		Primitive: typ,
		Schema:    schema,
	}
}

func Array(item *Type, sem NilSemantic, schema *oas.Schema) *Type {
	return &Type{
		Kind:        KindArray,
		Item:        item,
		Schema:      schema,
		NilSemantic: sem,
	}
}

func Alias(name string, to *Type) *Type {
	return &Type{
		Kind:       KindAlias,
		Name:       name,
		AliasTo:    to,
		Validators: to.Validators,
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

func Stream(name string) *Type {
	return &Type{
		Kind: KindStream,
		Name: name,
	}
}
