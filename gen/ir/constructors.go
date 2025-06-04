package ir

import (
	"path"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/jsonschema"
)

func Primitive(typ PrimitiveType, schema *jsonschema.Schema) *Type {
	return &Type{
		Kind:      KindPrimitive,
		Primitive: typ,
		Schema:    schema,
	}
}

func Array(item *Type, sem NilSemantic, schema *jsonschema.Schema) *Type {
	return &Type{
		Kind:        KindArray,
		Item:        item,
		Schema:      schema,
		NilSemantic: sem,
		Features:    item.CloneFeatures(),
	}
}

func Alias(name string, to *Type) *Type {
	return &Type{
		Kind:       KindAlias,
		Name:       name,
		AliasTo:    to,
		Validators: to.Validators,
		Features:   to.CloneFeatures(),
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

func Pointer(to *Type, sem NilSemantic) *Type {
	return &Type{
		Kind:        KindPointer,
		PointerTo:   to,
		NilSemantic: sem,
		Features:    to.CloneFeatures(),
	}
}

func Generic(name string, of *Type, v GenericVariant) *Type {
	name = v.Name() + name
	return &Type{
		Name:           name,
		Kind:           KindGeneric,
		GenericOf:      of,
		GenericVariant: v,
		Features:       of.CloneFeatures(),
	}
}

func Any(schema *jsonschema.Schema) *Type {
	return &Type{
		Kind:   KindAny,
		Schema: schema,
	}
}

func Stream(name string, schema *jsonschema.Schema) *Type {
	return &Type{
		Kind:   KindStream,
		Name:   name,
		Schema: schema,
	}
}

func External(schema *jsonschema.Schema) (*Type, error) {
	// If schema.XOgenType has no slashes or dots, it is a builtin type.
	if !strings.ContainsAny(schema.XOgenType, "/.") {
		return &Type{
			Kind:      KindPrimitive,
			Primitive: PrimitiveType(schema.XOgenType),
			Schema:    schema,
			External: ExternalType{
				GoName: schema.XOgenType,
			},
		}, nil
	}

	i := strings.LastIndex(schema.XOgenType, ".")
	if i < 0 || i == len(schema.XOgenType)-1 {
		return nil, errors.Errorf("%q is not a valid Go type (expected 'pkg/path.Type')", schema.XOgenType)
	}
	pkgPath := schema.XOgenType[:i]
	typeName := schema.XOgenType[i+1:]

	encode, decode, err := getExternalEncoding(pkgPath, typeName)
	if err != nil {
		return nil, errors.Wrapf(err, "get marshal kind for %q", schema.XOgenType)
	}

	return &Type{
		Kind:      KindPrimitive,
		Primitive: PrimitiveType(path.Base(pkgPath) + "." + typeName),
		Schema:    schema,
		External: ExternalType{
			PackagePath: pkgPath,
			GoName:      typeName,
			Encode:      encode,
			Decode:      decode,
		},
	}, nil
}
