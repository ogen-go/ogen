package ir

import (
	"fmt"
	"sort"

	ast "github.com/ogen-go/ogen/internal/ast2"
)

type Kind string

const (
	KindPrimitive Kind = "primitive"
	KindArray     Kind = "array"
	KindAlias     Kind = "alias"
	KindEnum      Kind = "enum"
	KindStruct    Kind = "struct"
	KindPointer   Kind = "pointer"
	KindInterface Kind = "interface"
	KindGeneric   Kind = "generic"
)

type Type struct {
	Kind             Kind
	Name             string              // only for struct, alias, interface, enum
	Primitive        PrimitiveType       // only for primitive, enum
	AliasTo          *Type               // only for alias
	PointerTo        *Type               // only for pointer
	Item             *Type               // only for array
	EnumValues       []interface{}       // only for enum
	Fields           []*Field            // only for struct
	Implements       map[*Type]struct{}  // only for struct, alias, enum
	Implementations  map[*Type]struct{}  // only for interface
	InterfaceMethods map[string]struct{} // only for interface
	Schema           *ast.Schema         // for all kinds except pointer, interface. Can be nil.
	NilSemantic      NilSemantic         // only for pointer
	GenericOf        *Type               // only for generic
	GenericVariant   GenericVariant      // only for generic
	Validators       Validators
}

// Format denotes whether custom formatting for Type is required while encoding
// or decoding.
//
// TODO(ernado): can we use t.JSON here?
func (t Type) Format() bool { return t.Primitive == Time }

// Tag of Field.
type Tag struct {
	JSON string // json tag, empty for none
}

type Field struct {
	Name string
	Type *Type
	Tag  Tag
}

func (t *Type) Is(vs ...Kind) bool {
	for _, v := range vs {
		if t.Kind == v {
			return true
		}
	}
	return false
}

func (t *Type) Implement(i *Type) {
	if !t.Is(KindStruct, KindAlias) || !i.Is(KindInterface) {
		panic("unreachable")
	}

	if t.Implements == nil {
		t.Implements = map[*Type]struct{}{}
	}

	i.Implementations[t] = struct{}{}
	t.Implements[i] = struct{}{}
}

func (t *Type) Unimplement(i *Type) {
	if !t.Is(KindStruct, KindAlias) || !i.Is(KindInterface) {
		panic("unreachable")
	}

	delete(i.Implementations, t)
	delete(t.Implements, i)
}

func (t *Type) AddMethod(name string) {
	if !t.Is(KindInterface) {
		panic("unreachable")
	}

	t.InterfaceMethods[name] = struct{}{}
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
		ms = t.InterfaceMethods
	case KindStruct, KindAlias, KindEnum:
		for i := range t.Implements {
			for m := range i.InterfaceMethods {
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
