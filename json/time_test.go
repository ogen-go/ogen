package json

import (
	"strconv"
	"testing"
	"time"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"
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

func BenchmarkEncodeDuration(b *testing.B) {
	t := time.Nanosecond +
		time.Microsecond +
		time.Millisecond +
		time.Second +
		time.Minute +
		time.Hour

	e := jx.GetEncoder()
	// Preallocate internal buffer.
	EncodeDuration(e, t)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		e.Reset()
		EncodeDuration(e, t)
	}
}

func TestEncodeDuration(t *testing.T) {
	tests := []time.Duration{
		0,
		10,
		time.Nanosecond,
		time.Microsecond,
		time.Millisecond,
		time.Second,
		time.Minute,
		time.Hour,
		time.Nanosecond +
			time.Microsecond +
			time.Millisecond +
			time.Second +
			time.Minute +
			time.Hour,
		// Tests from stdlib.
		1100 * time.Nanosecond,
		2200 * time.Microsecond,
		3300 * time.Millisecond,
		4*time.Minute + 5*time.Second,
		4*time.Minute + 5001*time.Millisecond,
		5*time.Hour + 6*time.Minute + 7001*time.Millisecond,
		8*time.Minute + 1*time.Nanosecond,
		1<<63 - 1,
		-1 << 63,
	}
	for _, tt := range tests {
		tt := tt
		expected := tt.String()
		t.Run(expected, func(t *testing.T) {
			e := jx.GetEncoder()
			EncodeDuration(e, tt)
			require.Equal(t, strconv.Quote(tt.String()), e.String())
		})
	}
}
