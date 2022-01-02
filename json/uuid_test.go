package json

import (
	"testing"

	"github.com/go-faster/jx"
	"github.com/google/uuid"
)

func BenchmarkDecodeUUID(b *testing.B) {
	e := jx.GetEncoder()
	defer jx.PutEncoder(e)

	u, err := uuid.NewUUID()
	if err != nil {
		b.Fatal(err)
	}
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
	e := jx.GetEncoder()
	defer jx.PutEncoder(e)

	u, err := uuid.NewUUID()
	if err != nil {
		b.Fatal(err)
	}

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
