package ir

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/internal/oas"
)

func TestJSONFields_RequiredMask(t *testing.T) {
	required := &Field{
		Spec: &oas.Property{
			Required: true,
		},
	}
	optional := &Field{
		Spec: &oas.Property{
			Required: false,
		},
	}
	fields16 := JSONFields{
		optional, required, required, required,
		required, optional, optional, required,
		optional, required, optional, required,
		optional, optional, optional, optional,
	}
	var (
		fields8Mask1 uint8 = 0b1001_1110
		fields8Mask2 uint8 = 0b0000_1010
	)
	tests := []struct {
		name  string
		j     JSONFields
		wantR []uint8
	}{
		{"Empty", nil, []uint8{0}},
		{"OneRequiredField", JSONFields{required}, []uint8{1}},
		{"OneOptionalField", JSONFields{optional}, []uint8{0}},
		{"OptionalRequired", JSONFields{optional, required}, []uint8{0b10}},
		{"RequiredRequired", JSONFields{required, required}, []uint8{0b11}},
		{"RequiredOptionalRequired", JSONFields{required, optional, required}, []uint8{0b101}},
		{"Fields16", fields16, []uint8{fields8Mask1, fields8Mask2}},
		{"ManyFields", func() (r JSONFields) {
			r = append(r, fields16...)
			r = append(r, fields16...)
			r = append(r, fields16...)
			return r
		}(), []uint8{
			fields8Mask1, fields8Mask2,
			fields8Mask1, fields8Mask2,
			fields8Mask1, fields8Mask2,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.j.RequiredMask()
			require.Equalf(t, tt.wantR, r, "%08b != %08b", tt.wantR[0], r[0])
		})
	}
}
