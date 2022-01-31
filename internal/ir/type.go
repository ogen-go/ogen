package ir

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/ogen-go/ogen/internal/oas"
)

type Kind string

const (
	KindPrimitive Kind = "primitive"
	KindArray     Kind = "array"
	KindMap       Kind = "map"
	KindAlias     Kind = "alias"
	KindEnum      Kind = "enum"
	KindStruct    Kind = "struct"
	KindPointer   Kind = "pointer"
	KindInterface Kind = "interface"
	KindGeneric   Kind = "generic"
	KindSum       Kind = "sum"
	KindStream    Kind = "stream"
)

type SumSpecMap struct {
	Key  string
	Type string
}

type TypeDiscriminator struct {
	Kind      Kind
	Primitive PrimitiveType
}

func (t TypeDiscriminator) Less(other TypeDiscriminator) bool {
	return t.Kind < other.Kind && t.Primitive < other.Primitive
}

func (t *TypeDiscriminator) Set(s *Type) {
	switch s.Kind {
	case KindPrimitive, KindEnum:
		// Treat enum as primitive.
		t.Kind = KindPrimitive
		t.Primitive = s.Primitive
	case KindArray:
		t.Kind = KindArray
	case KindAlias:
		t.Set(s.AliasTo)
	case KindPointer:
		t.Set(s.PointerTo)
	case KindGeneric:
		t.Set(s.GenericOf)
	default:
		// Treat all other types as struct.
		t.Kind = KindStruct
	}
}

// SumSpec for KindSum.
type SumSpec struct {
	Unique []*Field
	// DefaultMapping is name of default mapping.
	//
	// Used for variant which has no unique fields.
	DefaultMapping string

	// Discriminator is field name of sum type discriminator.
	Discriminator string
	// Mapping is discriminator value -> variant mapping.
	Mapping []SumSpecMap

	// TypeDiscriminator denotes to distinguish variants by type.
	TypeDiscriminator bool
}

type Type struct {
	Doc              string              // ogen documentation
	Kind             Kind                // kind
	Name             string              // only for struct, alias, interface, enum
	Primitive        PrimitiveType       // only for primitive, enum
	AliasTo          *Type               // only for alias
	PointerTo        *Type               // only for pointer
	SumOf            []*Type             // only for sum
	SumSpec          SumSpec             // only for sum
	Item             *Type               // only for array, map
	EnumVariants     []*EnumVariant      // only for enum
	Fields           []*Field            // only for struct
	Implements       map[*Type]struct{}  // only for struct, alias, enum
	Implementations  map[*Type]struct{}  // only for interface
	InterfaceMethods map[string]struct{} // only for interface
	Schema           *oas.Schema         // for all kinds except pointer, interface. Can be nil.
	NilSemantic      NilSemantic         // only for pointer
	GenericOf        *Type               // only for generic
	GenericVariant   GenericVariant      // only for generic
	Validators       Validators

	// Features contains a set of features the type must implement.
	// Available features: 'json', 'uri'.
	//
	// If some of these features are set, generator
	// generates additional encoding methods if needed.
	Features []string
}

func (t Type) GoDoc() []string {
	if t.Schema == nil {
		return nil
	}
	return prettyDoc(t.Schema.Description)
}

func (t Type) String() string {
	var b strings.Builder
	b.WriteString(string(t.Kind))
	b.WriteRune('(')
	b.WriteString(t.Go())
	b.WriteRune(')')
	return b.String()
}

type EnumVariant struct {
	Name  string
	Value interface{}
}

func (v *EnumVariant) ValueGo() string {
	switch v := v.Value.(type) {
	case string:
		return `"` + v + `"`
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (t *Type) Pointer(sem NilSemantic) *Type {
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
	return t.Primitive == Time
}

// Tag of Field.
type Tag struct {
	JSON string // json tag, empty for none
}

// EscapedJSON returns quoted and escaped JSON tag.
func (t Tag) EscapedJSON() string {
	return strconv.Quote(t.JSON)
}

// Field of structure.
type Field struct {
	Name string
	Type *Type
	Tag  Tag
	Spec *oas.Property
}

func (f Field) GoDoc() []string {
	if f.Spec == nil {
		return nil
	}
	return prettyDoc(f.Spec.Description)
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
	if !t.Is(KindStruct, KindMap, KindAlias, KindSum, KindStream) || !i.Is(KindInterface) {
		panic("unreachable")
	}

	if t.Implements == nil {
		t.Implements = map[*Type]struct{}{}
	}

	i.Implementations[t] = struct{}{}
	t.Implements[i] = struct{}{}
}

func (t *Type) Unimplement(i *Type) {
	if !t.Is(KindStruct, KindMap, KindAlias, KindSum) || !i.Is(KindInterface) {
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

func (t *Type) Go() string {
	switch t.Kind {
	case KindPrimitive:
		return t.Primitive.String()
	case KindArray:
		return "[]" + t.Item.Go()
	case KindPointer:
		return "*" + t.PointerTo.Go()
	case KindStruct, KindMap, KindAlias, KindInterface, KindGeneric, KindEnum, KindSum, KindStream:
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
	case KindStruct, KindMap, KindAlias, KindEnum, KindGeneric, KindSum, KindStream:
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

func (t *Type) ListImplementations() []*Type {
	if !t.Is(KindInterface) {
		panic("unreachable")
	}

	result := make([]*Type, 0, len(t.Implementations))
	for impl := range t.Implementations {
		result = append(result, impl)
	}
	sort.SliceStable(result, func(i, j int) bool {
		return strings.Compare(result[i].Name, result[j].Name) < 0
	})
	return result
}
