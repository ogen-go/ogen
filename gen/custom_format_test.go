package gen

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/gen/ir"
)

func typeOf[T any]() reflect.Type {
	return reflect.TypeOf(new(T)).Elem()
}

type foo struct{}

type ExportedFunc func() foo

func Test_checkImportableType(t *testing.T) {
	tests := []struct {
		typ     reflect.Type
		wantErr bool
	}{
		// Primitive types.
		{typeOf[bool](), false},
		{typeOf[int](), false},
		{typeOf[int8](), false},
		{typeOf[int16](), false},
		{typeOf[int32](), false},
		{typeOf[int64](), false},
		{typeOf[uint](), false},
		{typeOf[uint8](), false},
		{typeOf[uint16](), false},
		{typeOf[uint32](), false},
		{typeOf[uint64](), false},
		{typeOf[uintptr](), false},
		{typeOf[float32](), false},
		{typeOf[float64](), false},
		{typeOf[complex64](), false},
		{typeOf[complex128](), false},
		{typeOf[string](), false},
		{typeOf[unsafe.Pointer](), false},
		// Exported types.
		{typeOf[Generator](), false},
		{typeOf[ir.Kind](), false},
		{typeOf[ir.CustomFormat](), false},
		{typeOf[ExportedFunc](), false},

		// Negative cases.
		//
		// Unexported type.
		{typeOf[foo](), true},
		{typeOf[func(foo)](), true},
		{typeOf[func() foo](), true},
		{typeOf[*foo](), true},
		{typeOf[chan foo](), true},
		{typeOf[[]foo](), true},
		{typeOf[map[string]foo](), true},
		// Unnamed type.
		{typeOf[struct{}](), true},
		{typeOf[func(struct{})](), true},
		{typeOf[struct{ X int }](), true},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			check := require.Error
			if !tt.wantErr {
				check = require.NoError
			}

			err := checkImportableType(tt.typ)
			defer func() {
				t.Logf("Kind: %q", tt.typ.Kind())
				t.Logf("Package: %q", tt.typ.PkgPath())
				t.Logf("Name: %q", tt.typ.Name())
				if err != nil {
					t.Logf("Error: %v", err)
				}
			}()

			check(t, err)
		})
	}
}
