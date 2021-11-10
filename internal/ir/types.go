package ir

import (
	"fmt"

	"github.com/ogen-go/ogen/internal/oas"
	"github.com/ogen-go/ogen/validate"
)

type TypePrimitive struct {
	Type             PrimitiveType
	Schema           *oas.Schema
	StringValidation validate.String
	IntValidation    validate.Int
}

func (t *TypePrimitive) Go() string { return t.Type.String() }

type TypeArray struct {
	Item       *Type
	Semantic   NilSemantic
	Schema     *oas.Schema
	Validation validate.Array
}

func (t *TypeArray) Go() string { return "[]" + t.Item.Go() }

type TypeAlias struct {
	Name       string
	To         *Type
	Implements map[*TypeInterface]struct{}
}

func (t *TypeAlias) Go() string        { return t.Name }
func (t *TypeAlias) Methods() []string { return methods(t.Implements) }
func (t *TypeAlias) Implement(iface *TypeInterface) {
	if t.Implements == nil {
		t.Implements = map[*TypeInterface]struct{}{}
	}
	t.Implements[iface] = struct{}{}
	iface.Implementations[t] = struct{}{}
}

func (t *TypeAlias) Unimplement(iface *TypeInterface) {
	delete(t.Implements, iface)
	delete(iface.Implementations, t)
}

type TypeEnum struct {
	Name             string
	Type             PrimitiveType
	EnumVariants     []*EnumVariant
	StringValidation validate.String
	IntValidation    validate.Int
	Schema           *oas.Schema
}

func (t *TypeEnum) Go() string { return t.Name }

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

type TypeStruct struct {
	Name       string
	Fields     []*Field
	Implements map[*TypeInterface]struct{}
	Schema     *oas.Schema
}

func (t *TypeStruct) Go() string        { return t.Name }
func (t *TypeStruct) Methods() []string { return methods(t.Implements) }
func (t *TypeStruct) Implement(iface *TypeInterface) {
	if t.Implements == nil {
		t.Implements = map[*TypeInterface]struct{}{}
	}
	t.Implements[iface] = struct{}{}
	iface.Implementations[t] = struct{}{}
}

func (t *TypeStruct) Unimplement(iface *TypeInterface) {
	delete(t.Implements, iface)
	delete(iface.Implementations, t)
}

// Field of structure.
type Field struct {
	Name string
	Type *Type
	Tag  Tag
	Spec *oas.Property
}

// Tag of Field.
type Tag struct {
	JSON string // json tag, empty for none
}

type TypePointer struct {
	To       *Type
	Semantic NilSemantic
}

func (t *TypePointer) Go() string { return "*" + t.To.Go() }

type TypeInterface struct {
	Name            string
	Methods         map[string]struct{}
	Implementations map[Implementer]struct{}
}

func (t *TypeInterface) Go() string { return t.Name }

type Implementer interface {
	Implement(*TypeInterface)
	Unimplement(*TypeInterface)
}

func (t *TypeInterface) AddMethod(name string) {
	t.Methods[name] = struct{}{}
}

type TypeGeneric struct {
	Name    string
	Of      *Type
	Variant GenericVariant
}

func (t *TypeGeneric) Go() string { return t.Name }

type TypeSum struct {
	Name       string
	SumOf      []*Type
	SumSpec    SumSpec
	Implements map[*TypeInterface]struct{}
	Schema     *oas.Schema
}

func (t *TypeSum) Methods() []string { return methods(t.Implements) }

type SumSpecMap struct {
	Key  string
	Type string
}

// SumSpec for KindSum.
type SumSpec struct {
	Unique        []*Field
	Discriminator string
	Mapping       []SumSpecMap
}

func (t *TypeSum) Go() string { return t.Name }

func (t *TypeSum) Implement(iface *TypeInterface) {
	if t.Implements == nil {
		t.Implements = map[*TypeInterface]struct{}{}
	}
	t.Implements[iface] = struct{}{}
	iface.Implementations[t] = struct{}{}
}

func (t *TypeSum) Unimplement(iface *TypeInterface) {
	delete(t.Implements, iface)
	delete(iface.Implementations, t)
}

type TypeStream struct {
	Name       string
	Implements map[*TypeInterface]struct{}
}

func (t *TypeStream) Go() string        { return t.Name }
func (t *TypeStream) Methods() []string { return methods(t.Implements) }
func (t *TypeStream) Implement(iface *TypeInterface) {
	if t.Implements == nil {
		t.Implements = map[*TypeInterface]struct{}{}
	}
	t.Implements[iface] = struct{}{}
	iface.Implementations[t] = struct{}{}
}

func (t *TypeStream) Unimplement(iface *TypeInterface) {
	delete(t.Implements, iface)
	delete(iface.Implementations, t)
}
