package json

import (
	"testing"
	"time"

	"github.com/go-faster/jx"
)

func BenchmarkEncodeDate(b *testing.B) {
	t := time.Now()
	e := jx.GetEncoder()
	// Preallocate internal buffer.
	EncodeDate(e, t)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		e.Reset()
		EncodeDate(e, t)
	}
}

func BenchmarkEncodeTime(b *testing.B) {
	t := time.Now()
	e := jx.GetEncoder()
	// Preallocate internal buffer.
	EncodeTime(e, t)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		e.Reset()
		EncodeTime(e, t)
	}
}

func BenchmarkEncodeDateTime(b *testing.B) {
	t := time.Now()
	e := jx.GetEncoder()
	// Preallocate internal buffer.
	EncodeDateTime(e, t)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		e.Reset()
		EncodeDateTime(e, t)
	}
}

