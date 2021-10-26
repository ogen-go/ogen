package ir

import (
	"fmt"
	"sort"

	ast "github.com/ogen-go/ogen/internal/ast2"
)

type TypeKind string

const (
	KindPrimitive TypeKind = "primitive"
	KindArray     TypeKind = "array"
	KindAlias     TypeKind = "alias"
	KindEnum      TypeKind = "enum"
	KindStruct    TypeKind = "struct"
	KindPointer   TypeKind = "pointer"
	KindInterface TypeKind = "interface"
)

type Type struct {
	Kind            TypeKind
	Name            string              // only for struct, alias, interface, enum
	Primitive       PrimitiveType       // only for primitive, enum
	AliasTo         *Type               // only for alias
	PointerTo       *Type               // only for pointer
	Item            *Type               // only for array
	EnumValues      []interface{}       // only for enum
	Fields          []*StructField      // only for struct
	Implements      map[*Type]struct{}  // only for struct, alias, enum
	Implementations map[*Type]struct{}  // only for interface
	IfaceMethods    map[string]struct{} // only for interface
	Spec            *ast.Schema         // for all kinds except pointer, interface. Can be nil.
}

type StructField struct {
	Name string
	Type *Type
	Tag  string
}

func (t *Type) Is(vs ...TypeKind) bool {
	for _, v := range vs {
		if t.Kind == v {
			return true
		}
	}
	return false
}

func (t *Type) Implement(iface *Type) {
	if !t.Is(KindStruct, KindAlias) || !iface.Is(KindInterface) {
		panic("unreachable")
	}

	if t.Implements == nil {
		t.Implements = map[*Type]struct{}{}
	}

	iface.Implementations[t] = struct{}{}
	t.Implements[iface] = struct{}{}
}

func (t *Type) Unimplement(iface *Type) {
	if !t.Is(KindStruct, KindAlias) || !iface.Is(KindInterface) {
		panic("unreachable")
	}

	delete(iface.Implementations, t)
	delete(t.Implements, iface)
}

func (t *Type) AddMethod(name string) {
	if !t.Is(KindInterface) {
		panic("unreachable")
	}

	t.IfaceMethods[name] = struct{}{}
}

func (t *Type) GoType() string {
	switch t.Kind {
	case KindPrimitive, KindEnum:
		return t.Primitive.String()
	case KindArray:
		return "[]" + t.Item.GoType()
	case KindPointer:
		return "*" + t.PointerTo.GoType()
	case KindStruct, KindAlias, KindInterface:
		return t.Name
	default:
		panic(fmt.Sprintf("unexpected kind: %s", t.Kind))
	}
}

func (t *Type) Methods() []string {
	ms := make(map[string]struct{})
	switch t.Kind {
	case KindInterface:
		ms = t.IfaceMethods
	case KindStruct, KindAlias, KindEnum:
		for iface := range t.Implements {
			for m := range iface.IfaceMethods {
				ms[m] = struct{}{}
			}
		}
	default:
		panic(fmt.Sprintf("unexpected kind: %s", t.Kind))
	}

	var result []string
	for m := range ms {
		result = append(result, m)
	}
	sort.Strings(result)
	return result
}
