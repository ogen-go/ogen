package json

import (
	"net/netip"
	"testing"

	"github.com/go-faster/jx"
)

func BenchmarkEncodeIP(b *testing.B) {
	ip := netip.MustParseAddr("127.0.0.1")
	e := jx.GetEncoder()
	// Preallocate internal buffer.
	EncodeIP(e, ip)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		e.Reset()
		EncodeIP(e, ip)
	}
}
