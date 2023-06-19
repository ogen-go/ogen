package json

import (
	"fmt"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"
)

func testDecodeEncode[T any](
	decode func(d *jx.Decoder) (T, error),
	encode func(e *jx.Encoder, v T),
	input string,
	wantVal T,
	wantErr bool,
) func(t *testing.T) {
	return func(t *testing.T) {
		a := require.New(t)
		d := jx.DecodeStr(input)

		got, err := decode(d)
		if wantErr {
			a.Errorf(err, "input: %q", input)
			return
		}
		a.NoError(err)
		a.Equal(wantVal, got)

		e := jx.GetEncoder()
		encode(e, wantVal)

		d.ResetBytes(e.Bytes())
		got2, err := decode(d)
		a.NoError(err)
		a.Equal(wantVal, got2)
	}
}

func TestStringInt8(t *testing.T) {
	tests := []struct {
		input   string
		wantVal int8
		wantErr bool
	}{
		{`"-1"`, -1, false},
		{`"0"`, 0, false},
		{`"1"`, 1, false},
		{`"10"`, 10, false},
		{`"100"`, 100, false},

		{"1", 0, true},
		{`"1000"`, 0, true},
		{`"foo"`, 0, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), testDecodeEncode(
			DecodeStringInt8,
			EncodeStringInt8,
			tt.input,
			tt.wantVal,
			tt.wantErr,
		))
	}
}

func TestStringInt16(t *testing.T) {
	tests := []struct {
		input   string
		wantVal int16
		wantErr bool
	}{
		{`"-1"`, -1, false},
		{`"0"`, 0, false},
		{`"1"`, 1, false},
		{`"10"`, 10, false},
		{`"10000"`, 10000, false},

		{"1", 0, true},
		{`"100000"`, 0, true},
		{`"foo"`, 0, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), testDecodeEncode(
			DecodeStringInt16,
			EncodeStringInt16,
			tt.input,
			tt.wantVal,
			tt.wantErr,
		))
	}
}

func TestStringInt32(t *testing.T) {
	tests := []struct {
		input   string
		wantVal int32
		wantErr bool
	}{
		{`"-1"`, -1, false},
		{`"0"`, 0, false},
		{`"1"`, 1, false},
		{`"10"`, 10, false},
		{`"10000"`, 10000, false},

		{"1", 0, true},
		{`"100000000000"`, 0, true},
		{`"foo"`, 0, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), testDecodeEncode(
			DecodeStringInt32,
			EncodeStringInt32,
			tt.input,
			tt.wantVal,
			tt.wantErr,
		))
	}
}

func TestStringInt64(t *testing.T) {
	tests := []struct {
		input   string
		wantVal int64
		wantErr bool
	}{
		{`"-1"`, -1, false},
		{`"0"`, 0, false},
		{`"1"`, 1, false},
		{`"10"`, 10, false},
		{`"10000"`, 10000, false},

		{"1", 0, true},
		{`"100000000000000000000000000"`, 0, true},
		{`"foo"`, 0, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), testDecodeEncode(
			DecodeStringInt64,
			EncodeStringInt64,
			tt.input,
			tt.wantVal,
			tt.wantErr,
		))
	}
}

func BenchmarkDecodeStringInt32(b *testing.B) {
	var (
		d     = jx.GetDecoder()
		input = []byte(`"1234567890"`)
		val   int32
		err   error
	)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		d.ResetBytes(input)
		val, err = DecodeStringInt32(d)
	}

	if val != 1234567890 || err != nil {
		b.Fatal(val, err)
	}
}

func BenchmarkDecodeStringInt64(b *testing.B) {
	var (
		d     = jx.GetDecoder()
		input = []byte(`"1234567890"`)
		val   int64
		err   error
	)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		d.ResetBytes(input)
		val, err = DecodeStringInt64(d)
	}

	if val != 1234567890 || err != nil {
		b.Fatal(val, err)
	}
}
