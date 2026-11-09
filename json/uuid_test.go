package json

import (
	"testing"
	"uuid"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"
)

func BenchmarkDecodeUUID(b *testing.B) {
	e := &jx.Encoder{}

	u := uuid.New()
	EncodeUUID(e, u)
	data := e.Bytes()

	d := jx.GetDecoder()
	defer jx.PutDecoder(d)

	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		d.ResetBytes(data)
		if _, err := DecodeUUID(d); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeUUID(b *testing.B) {
	e := &jx.Encoder{}

	u := uuid.New()

	EncodeUUID(e, u)
	data := e.Bytes()

	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		e.Reset()
		EncodeUUID(e, u)
	}
}

func Test_hexEncode(t *testing.T) {
	a := require.New(t)

	u := uuid.New()

	var dst [36]byte
	hexEncode(&dst, u)
	a.Equal(u.String(), string(dst[:]))
}
