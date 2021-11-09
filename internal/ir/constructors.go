package ir

import "github.com/ogen-go/ogen/internal/oas"

func Primitive(typ PrimitiveType, schema *oas.Schema) *Type {
	return &Type{
		Kind: KindPrimitive,
		Primitive: &TypePrimitive{
			Type:   typ,
			Schema: schema,
		},
	}
}

func Array(item *Type, sem NilSemantic, schema *oas.Schema) *Type {
	return &Type{
		Kind: KindArray,
		Array: &TypeArray{
			Item:     item,
			Semantic: sem,
			Schema:   schema,
		},
	}
}

func Alias(name string, to *Type) *Type {
	return &Type{
		Kind: KindAlias,
		Alias: &TypeAlias{
			Name: name,
			To:   to,
		},
	}
}

func Interface(name string) *Type {
	return &Type{
		Kind: KindInterface,
		Interface: &TypeInterface{
			Name:            name,
			Implementations: map[Implementer]struct{}{},
			Methods:         map[string]struct{}{},
		},
	}
}

func Pointer(typ *Type, sem NilSemantic) *Type {
	return &Type{
		Kind: KindPointer,
		Pointer: &TypePointer{
			To:       typ,
			Semantic: sem,
		},
	}
}

func Stream() *Type {
	return &Type{
		Kind: KindStream,
		Stream: &TypeStream{
			Name: "Stream",
		},
	}
}
