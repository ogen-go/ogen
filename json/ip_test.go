package json

import (
	"fmt"
	"net/netip"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"
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

func TestIPv4(t *testing.T) {
	addr := netip.MustParseAddr
	tests := []struct {
		input       string
		want        netip.Addr
		errContains string
	}{
		// Valid IPv4.
		{`"1.1.1.1"`, addr("1.1.1.1"), ""},
		{`"127.0.0.1"`, addr("127.0.0.1"), ""},

		// Invalid IPv4.
		{`"127.0.0"`, netip.Addr{}, "bad ip format"},
		{`"127.0.0.-1"`, netip.Addr{}, "bad ip format"},
		{`"127.256.0.0"`, netip.Addr{}, "bad ip format"},

		// Wrong IP version.
		{`"::1"`, netip.Addr{}, "wrong ip version"},
		{`"2001:db8::1"`, netip.Addr{}, "wrong ip version"},

		// Wrong type.
		{`1`, netip.Addr{}, "start: unexpected byte"},
		{`true`, netip.Addr{}, "start: unexpected byte"},
		{`null`, netip.Addr{}, "start: unexpected byte"},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			d := jx.DecodeStr(tt.input)
			got, err := DecodeIPv4(d)
			if tt.errContains != "" {
				a.ErrorContains(err, tt.errContains)
				return
			}
			a.NoError(err)
			a.Equal(tt.want, got)

			e := jx.GetEncoder()
			EncodeIPv4(e, got)
			got2, err := DecodeIPv4(jx.DecodeBytes(e.Bytes()))
			a.NoError(err)
			a.True(got.Compare(got2) == 0)
		})
	}
}

func TestIPv6(t *testing.T) {
	addr := netip.MustParseAddr
	tests := []struct {
		input       string
		want        netip.Addr
		errContains string
	}{
		// Valid IPv6.
		{`"::1"`, addr("::1"), ""},
		{`"2001:db8::1"`, addr("2001:db8::1"), ""},

		// Invalid IPv6.
		{`":1"`, netip.Addr{}, "bad ip format"},
		{`"::-1"`, netip.Addr{}, "bad ip format"},
		{`"2001:hb8::1"`, netip.Addr{}, "bad ip format"},

		// Wrong IP version.
		{`"127.0.0.1"`, netip.Addr{}, "wrong ip version"},
		{`"1.1.1.1"`, netip.Addr{}, "wrong ip version"},

		// Wrong type.
		{`1`, netip.Addr{}, "start: unexpected byte"},
		{`true`, netip.Addr{}, "start: unexpected byte"},
		{`null`, netip.Addr{}, "start: unexpected byte"},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			d := jx.DecodeStr(tt.input)
			got, err := DecodeIPv6(d)
			if tt.errContains != "" {
				a.ErrorContains(err, tt.errContains)
				return
			}
			a.NoError(err)
			a.Equal(tt.want, got)

			e := jx.GetEncoder()
			EncodeIPv6(e, got)
			got2, err := DecodeIPv6(jx.DecodeBytes(e.Bytes()))
			a.NoError(err)
			a.True(got.Compare(got2) == 0)
		})
	}
}
