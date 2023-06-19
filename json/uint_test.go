package json

import (
	"fmt"
	"testing"
)

func TestStringUint8(t *testing.T) {
	tests := []struct {
		input   string
		wantVal uint8
		wantErr bool
	}{
		{`"0"`, 0, false},
		{`"1"`, 1, false},
		{`"10"`, 10, false},
		{`"100"`, 100, false},

		{"1", 0, true},
		{`"100000000000"`, 0, true},
		{`"foo"`, 0, true},
		{`"-1"`, 0, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), testDecodeEncode(
			DecodeStringUint8,
			EncodeStringUint8,
			tt.input,
			tt.wantVal,
			tt.wantErr,
		))
	}
}

func TestStringUint16(t *testing.T) {
	tests := []struct {
		input   string
		wantVal uint16
		wantErr bool
	}{
		{`"0"`, 0, false},
		{`"1"`, 1, false},
		{`"10"`, 10, false},
		{`"10000"`, 10000, false},

		{"1", 0, true},
		{`"100000000000"`, 0, true},
		{`"foo"`, 0, true},
		{`"-1"`, 0, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), testDecodeEncode(
			DecodeStringUint16,
			EncodeStringUint16,
			tt.input,
			tt.wantVal,
			tt.wantErr,
		))
	}
}

func TestStringUint32(t *testing.T) {
	tests := []struct {
		input   string
		wantVal uint32
		wantErr bool
	}{
		{`"0"`, 0, false},
		{`"1"`, 1, false},
		{`"10"`, 10, false},
		{`"10000"`, 10000, false},

		{"1", 0, true},
		{`"100000000000"`, 0, true},
		{`"foo"`, 0, true},
		{`"-1"`, 0, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), testDecodeEncode(
			DecodeStringUint32,
			EncodeStringUint32,
			tt.input,
			tt.wantVal,
			tt.wantErr,
		))
	}
}

func TestStringUint64(t *testing.T) {
	tests := []struct {
		input   string
		wantVal uint64
		wantErr bool
	}{
		{`"0"`, 0, false},
		{`"1"`, 1, false},
		{`"10"`, 10, false},
		{`"10000"`, 10000, false},

		{"1", 0, true},
		{`"100000000000000000000000000"`, 0, true},
		{`"foo"`, 0, true},
		{`"-1"`, 0, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), testDecodeEncode(
			DecodeStringUint64,
			EncodeStringUint64,
			tt.input,
			tt.wantVal,
			tt.wantErr,
		))
	}
}
