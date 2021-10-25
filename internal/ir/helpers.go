package ir

import ast "github.com/ogen-go/ogen/internal/ast2"

func Primitive(typ PrimitiveType, spec *ast.Schema) *Type {
	return &Type{
		Kind:      KindPrimitive,
		Primitive: typ,
		Spec:      spec,
	}
}

func Array(item *Type, spec *ast.Schema) *Type {
	return &Type{
		Kind: KindArray,
		Item: item,
		Spec: spec,
	}
}

func Alias(name string, to *Type) *Type {
	return &Type{
		Kind:    KindAlias,
		Name:    name,
		AliasTo: to,
	}
}

func Iface(name string) *Type {
	return &Type{
		Name:            name,
		Kind:            KindInterface,
		IfaceMethods:    map[string]struct{}{},
		Implementations: map[*Type]struct{}{},
	}
}
