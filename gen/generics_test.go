package gen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/gen/ir"
)

func TestBoxTypeEmptyStruct(t *testing.T) {
	newEmpty := func() *ir.Type {
		return &ir.Type{Kind: ir.KindStruct, Name: "Empty"}
	}
	newFilled := func() *ir.Type {
		return &ir.Type{
			Kind: ir.KindStruct,
			Name: "Filled",
			Fields: []*ir.Field{
				{Name: "A", Type: ir.Primitive(ir.String, nil)},
			},
		}
	}

	tests := []struct {
		name     string
		typ      func() *ir.Type
		variant  ir.GenericVariant
		wantKind ir.Kind
		wantName string
	}{
		{
			name:     "EmptyRequired",
			typ:      newEmpty,
			variant:  ir.GenericVariant{},
			wantKind: ir.KindStruct,
			wantName: "Empty",
		},
		{
			// An optional empty struct stays a pointer, since there is no null to track.
			name:     "EmptyOptional",
			typ:      newEmpty,
			variant:  ir.GenericVariant{Optional: true},
			wantKind: ir.KindPointer,
		},
		{
			name:     "EmptyNullable",
			typ:      newEmpty,
			variant:  ir.GenericVariant{Nullable: true},
			wantKind: ir.KindGeneric,
			wantName: "NilEmpty",
		},
		{
			name:     "EmptyNullableOptional",
			typ:      newEmpty,
			variant:  ir.GenericVariant{Nullable: true, Optional: true},
			wantKind: ir.KindGeneric,
			wantName: "OptNilEmpty",
		},
		{
			name:     "FilledNullable",
			typ:      newFilled,
			variant:  ir.GenericVariant{Nullable: true},
			wantKind: ir.KindGeneric,
			wantName: "NilFilled",
		},
		{
			name:     "FilledNullableOptional",
			typ:      newFilled,
			variant:  ir.GenericVariant{Nullable: true, Optional: true},
			wantKind: ir.KindGeneric,
			wantName: "OptNilFilled",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := require.New(t)

			got, err := boxType(tt.typ(), tt.variant)
			a.NoError(err)
			a.Equal(tt.wantKind, got.Kind)
			if tt.wantName != "" {
				a.Equal(tt.wantName, got.Name)
			}
			if got.IsGeneric() {
				a.Equal(tt.variant, got.GenericVariant)
			}
		})
	}
}

func TestBoxTypeEmptyStructAlias(t *testing.T) {
	a := require.New(t)

	alias := &ir.Type{
		Kind:    ir.KindAlias,
		Name:    "Alias",
		AliasTo: &ir.Type{Kind: ir.KindStruct, Name: "Empty"},
	}

	got, err := boxType(alias, ir.GenericVariant{Nullable: true})
	a.NoError(err)
	a.True(got.IsGeneric())
	a.Equal("NilAlias", got.Name)
}
