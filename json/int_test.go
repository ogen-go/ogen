package json

import (
	"fmt"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"
)

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
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)
			d := jx.DecodeStr(tt.input)

			got, err := DecodeStringInt32(d)
			if tt.wantErr {
				a.Error(err)
				return
			}
			a.NoError(err)
			a.Equal(tt.wantVal, got)

			e := jx.GetEncoder()
			EncodeStringInt32(e, tt.wantVal)

			d.ResetBytes(e.Bytes())
			got2, err := DecodeStringInt32(d)
			a.NoError(err)
			a.Equal(tt.wantVal, got2)
		})
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
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)
			d := jx.DecodeStr(tt.input)

			got, err := DecodeStringInt64(d)
			if tt.wantErr {
				a.Error(err)
				return
			}
			a.NoError(err)
			a.Equal(tt.wantVal, got)

			e := jx.GetEncoder()
			EncodeStringInt64(e, tt.wantVal)

			d.ResetBytes(e.Bytes())
			got2, err := DecodeStringInt64(d)
			a.NoError(err)
			a.Equal(tt.wantVal, got2)
		})
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
