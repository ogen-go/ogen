package ir

import (
	"fmt"
	"sort"

	"github.com/ogen-go/ogen/internal/oas"
	"github.com/ogen-go/ogen/validate"
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
	KindSum       Kind = "sum"
	KindStream    Kind = "stream"
)

type Type struct {
	ghostName string
	Doc       string
	Kind      Kind
	Primitive *TypePrimitive
	Array     *TypeArray
	Enum      *TypeEnum
	Alias     *TypeAlias
	Struct    *TypeStruct
	Pointer   *TypePointer
	Interface *TypeInterface
	Generic   *TypeGeneric
	Sum       *TypeSum
	Stream    *TypeStream
}

func (t *Type) MakePointer(sem NilSemantic) *Type {
	return Pointer(t, sem)
}

// Format denotes whether custom formatting for Type is required while encoding
// or decoding.
//
// TODO(ernado): can we use t.JSON here?
func (t *Type) Format() bool {
	if t == nil {
		return false
	}
	return t.Kind == KindPrimitive && t.Primitive.Type == Time
}

// Tag of Field.
type Tag struct {
	JSON string // json tag, empty for none
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
	if !i.Is(KindInterface) {
		panic("unreachable")
	}

	switch t.Kind {
	case KindStruct:
		t.Struct.Implement(i.Interface)
	case KindAlias:
		t.Alias.Implement(i.Interface)
	case KindSum:
		t.Sum.Implement(i.Interface)
	case KindStream:
		t.Stream.Implement(i.Interface)
	default:
		panic(fmt.Sprintf("unexpected kind: %s", t.Kind))
	}
}

func (t *Type) Unimplement(i *Type) {
	if !i.Is(KindInterface) {
		panic("unreachable")
	}

	switch t.Kind {
	case KindStruct:
		t.Struct.Unimplement(i.Interface)
	case KindAlias:
		t.Alias.Unimplement(i.Interface)
	case KindSum:
		t.Sum.Unimplement(i.Interface)
	case KindStream:
		t.Stream.Unimplement(i.Interface)
	default:
		panic(fmt.Sprintf("unexpected kind: %s", t.Kind))
	}
}

func (t *Type) AddMethod(name string) {
	if !t.Is(KindInterface) {
		panic("unreachable")
	}

	t.Interface.AddMethod(name)
}

func (t *Type) Go() string {
	switch t.Kind {
	case KindPrimitive:
		return t.Primitive.Go()
	case KindArray:
		return t.Array.Go()
	case KindPointer:
		return t.Pointer.Go()
	case KindStruct:
		return t.Struct.Go()
	case KindAlias:
		return t.Alias.Go()
	case KindInterface:
		return t.Interface.Go()
	case KindGeneric:
		return t.Generic.Go()
	case KindEnum:
		return t.Enum.Go()
	case KindSum:
		return t.Sum.Go()
	case KindStream:
		return t.Stream.Go()
	default:
		panic(fmt.Sprintf("unexpected kind: %s", t.Kind))
	}
}

func (t *Type) Methods() []string {
	ms := make(map[string]struct{})
	if t.Is(KindInterface) {
		ms = t.Interface.Methods
	} else {
		var impl map[*TypeInterface]struct{}
		switch t.Kind {
		case KindStruct:
			impl = t.Struct.Implements
		case KindAlias:
			impl = t.Alias.Implements
		case KindSum:
			impl = t.Sum.Implements
		case KindStream:
			impl = t.Stream.Implements
		default:
			panic(fmt.Sprintf("unexpected kind: %s", t.Kind))
		}

		for i := range impl {
			for m := range i.Methods {
				ms[m] = struct{}{}
			}
		}
	}

	var result []string
	for m := range ms {
		result = append(result, m)
	}
	sort.Strings(result)
	return result
}

// func (t *Type) ListImplementations() []*Type {
// 	if !t.Is(KindInterface) {
// 		panic("unreachable")
// 	}

// 	result := make([]*Type, 0, len(t.Implementations))
// 	for impl := range t.Implementations {
// 		result = append(result, impl)
// 	}
// 	sort.SliceStable(result, func(i, j int) bool {
// 		return strings.Compare(result[i].Name, result[j].Name) < 0
// 	})
// 	return result
// }

func (t *Type) Schema() *oas.Schema {
	switch t.Kind {
	case KindPrimitive:
		return t.Primitive.Schema
	case KindArray:
		return t.Array.Schema
	case KindStruct:
		return t.Struct.Schema
	case KindSum:
		return t.Sum.Schema
	case KindEnum:
		return t.Enum.Schema
	default:
		return nil
	}
}

func (t *Type) Name() (string, bool) {
	switch t.Kind {
	case KindStruct:
		return t.Struct.Name, true
	case KindEnum:
		return t.Enum.Name, true
	case KindAlias:
		return t.Alias.Name, true
	case KindGeneric:
		return t.Generic.Name, true
	case KindSum:
		return t.Sum.Name, true
	case KindStream:
		return t.Stream.Name, true
	default:
		return "", false
	}
}

func (t *Type) SetIntValidation(v validate.Int) {
	if !t.IsNumeric() {
		panic("unreachable")
	}

	switch t.Kind {
	case KindPrimitive:
		t.Primitive.IntValidation = v
	case KindEnum:
		t.Enum.IntValidation = v
	default:
		panic("unreachable")
	}
}

func (t *Type) SetStringValidation(v validate.String) {
	switch t.Kind {
	case KindPrimitive:
		// if t.Primitive.Type != String {
		// 	panic(fmt.Sprintf("invalid primitive type: %q", t.Primitive.Type))
		// }
		t.Primitive.StringValidation = v
	case KindEnum:
		// if t.Enum.Type != String {
		// 	panic(fmt.Sprintf("invalid primitive type: %q", t.Enum.Type))
		// }
		t.Enum.StringValidation = v
	default:
		panic(fmt.Sprintf("invalid type kind: %q", t.Kind))
	}
}
