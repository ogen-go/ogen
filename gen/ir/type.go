package ir

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ogen-go/ogen/internal/naming"
	"github.com/ogen-go/ogen/jsonschema"
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
	KindAny       Kind = "any"
	KindStream    Kind = "stream"
)

type SumSpecMap struct {
	Key  string
	Type *Type
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
	Doc                 string              // ogen documentation
	Kind                Kind                // kind
	Name                string              // only for struct, alias, interface, enum, stream, generic, map, sum
	Primitive           PrimitiveType       // only for primitive, enum
	AliasTo             *Type               // only for alias
	PointerTo           *Type               // only for pointer
	SumOf               []*Type             // only for sum
	SumSpec             SumSpec             // only for sum
	Item                *Type               // only for array, map
	EnumVariants        []*EnumVariant      // only for enum
	Fields              []*Field            // only for struct
	Implements          map[*Type]struct{}  // only for struct, alias, enum
	Implementations     map[*Type]struct{}  // only for interface
	InterfaceMethods    map[string]struct{} // only for interface
	Schema              *jsonschema.Schema  // for all kinds except pointer, interface. Can be nil.
	NilSemantic         NilSemantic         // only for pointer
	GenericOf           *Type               // only for generic
	GenericVariant      GenericVariant      // only for generic
	MapPattern          *regexp.Regexp      // only for map
	DenyAdditionalProps bool                // only for map and struct
	Validators          Validators

	// Features contains a set of features the type must implement.
	// Available features: 'json', 'uri'.
	//
	// If some of these features are set, generator
	// generates additional encoding methods if needed.
	Features []string
}

// HasName indicates whether the type has a name.
func (t Type) HasName() (string, bool) {
	if t.Is(KindStruct, KindAlias, KindInterface, KindEnum, KindStream, KindGeneric, KindMap, KindSum) {
		return t.Name, true
	}

	return "", false
}

// GoDoc returns type godoc.
func (t Type) GoDoc() []string {
	s := t.Schema
	if s == nil {
		return nil
	}

	doc := s.Description
	if doc == "" {
		doc = s.Summary
	}

	var notice string
	if s.Deprecated {
		notice = "Deprecated: schema marks this type as deprecated."
	}
	return prettyDoc(doc, notice)
}

// Default returns default value of this type, if it is set.
func (t Type) Default() Default {
	schema := t.Schema
	if schema == nil {
		return Default{}
	}
	return Default{
		Value: schema.Default,
		Set:   schema.DefaultSet,
	}
}

func (t Type) String() string {
	var b strings.Builder
	b.WriteString(string(t.Kind))
	b.WriteRune('(')
	b.WriteString(t.Go())
	b.WriteRune(')')
	if s := t.Schema; s != nil && s.Ref != "" {
		b.WriteRune('(')
		b.WriteString(s.Ref)
		b.WriteRune(')')
	}
	return b.String()
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

func (t *Type) Is(vs ...Kind) bool {
	for _, v := range vs {
		if t.Kind == v {
			return true
		}
	}
	return false
}

// Go returns valid Go type for this Type.
func (t *Type) Go() string {
	switch t.Kind {
	case KindPrimitive:
		return t.Primitive.String()
	case KindAny:
		return "jx.Raw"
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

// NamePostfix returns name postfix for optional wrapper.
func (t *Type) NamePostfix() string {
	switch t.Kind {
	case KindPrimitive:
		if t.Primitive == Null {
			return "Null"
		}
		s := t.Schema
		typePrefix := func(f string) string {
			switch s.Type {
			case jsonschema.String:
				return "String" + naming.Capitalize(f)
			default:
				return f
			}
		}
		switch f := s.Format; f {
		case "uuid":
			return "UUID"
		case "date":
			return "Date"
		case "time":
			return "Time"
		case "date-time":
			return "DateTime"
		case "duration":
			return "Duration"
		case "ip":
			return "IP"
		case "ipv4":
			return "IPv4"
		case "ipv6":
			return "IPv6"
		case "uri":
			return "URI"
		case "int32", "int64":
			if s.Type != jsonschema.String {
				return t.Primitive.String()
			}
			return "String" + naming.Capitalize(f)
		case "unix", "unix-seconds":
			return typePrefix("UnixSeconds")
		case "unix-nano":
			return typePrefix("UnixNano")
		case "unix-micro":
			return typePrefix("UnixMicro")
		case "unix-milli":
			return typePrefix("UnixMilli")
		default:
			return t.Primitive.String()
		}
	case KindArray:
		return t.Item.NamePostfix() + "Array"
	case KindAny:
		return "Any"
	case KindPointer:
		return t.PointerTo.NamePostfix() + "Pointer"
	case KindStruct, KindMap, KindAlias, KindInterface, KindGeneric, KindEnum, KindSum, KindStream:
		return t.Name
	default:
		panic(fmt.Sprintf("unexpected kind: %s", t.Kind))
	}
}
