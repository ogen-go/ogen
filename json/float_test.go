package json

import (
	"fmt"
	"math"
	"testing"
)

func TestStringFloat32(t *testing.T) {
	tests := []struct {
		input   string
		wantVal float32
		wantErr bool
	}{
		{`"-1"`, -1, false},
		{`"0"`, 0, false},
		{`"0.0"`, 0, false},
		{`"1"`, 1, false},
		{`"10"`, 10, false},
		{`"1e1"`, 10, false},
		{`"100"`, 100, false},
		{`"inf"`, float32(math.Inf(0)), false},
		{`"infinity"`, float32(math.Inf(0)), false},
		{`"-inf"`, float32(math.Inf(-1)), false},
		{`"-infinity"`, float32(math.Inf(-1)), false},

		{"1", 0, true},
		{`"null"`, 0, true},
		{`"foo"`, 0, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), testDecodeEncode(
			DecodeStringFloat32,
			EncodeStringFloat32,
			tt.input,
			tt.wantVal,
			tt.wantErr,
		))
	}
}

func TestStringFloat64(t *testing.T) {
	tests := []struct {
		input   string
		wantVal float64
		wantErr bool
	}{
		{`"-1"`, -1, false},
		{`"0"`, 0, false},
		{`"0.0"`, 0, false},
		{`"1"`, 1, false},
		{`"10"`, 10, false},
		{`"1e1"`, 10, false},
		{`"100"`, 100, false},
		{`"inf"`, math.Inf(0), false},
		{`"infinity"`, math.Inf(0), false},
		{`"-inf"`, math.Inf(-1), false},
		{`"-infinity"`, math.Inf(-1), false},

		{"1", 0, true},
		{`"null"`, 0, true},
		{`"foo"`, 0, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), testDecodeEncode(
			DecodeStringFloat64,
			EncodeStringFloat64,
			tt.input,
			tt.wantVal,
			tt.wantErr,
		))
	}
}
