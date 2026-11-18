package ir

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTypeCanGeneric(t *testing.T) {
	filled := &Type{
		Kind:   KindStruct,
		Fields: []*Field{{Name: "A", Type: Primitive(String, nil)}},
	}
	empty := &Type{Kind: KindStruct}

	tests := []struct {
		name               string
		typ                *Type
		canGeneric         bool
		canGenericNullable bool
	}{
		{"Primitive", Primitive(String, nil), true, true},
		{"Struct", filled, true, true},
		{"EmptyStruct", empty, false, true},
		{"AliasOfStruct", &Type{Kind: KindAlias, AliasTo: filled}, true, true},
		{"AliasOfEmptyStruct", &Type{Kind: KindAlias, AliasTo: empty}, false, true},
		{"EmptySum", &Type{Kind: KindSum}, false, false},
		{"Map", &Type{Kind: KindMap}, true, true},
		{"Pointer", &Type{Kind: KindPointer, PointerTo: filled}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := require.New(t)
			a.Equal(tt.canGeneric, tt.typ.CanGeneric())
			a.Equal(tt.canGenericNullable, tt.typ.CanGenericNullable())
		})
	}
}
