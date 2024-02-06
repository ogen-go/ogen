package json

import (
	"fmt"
	"net"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"
)

func BenchmarkEncodeMAC(b *testing.B) {
	mac, err := net.ParseMAC("00:11:22:33:44:55")
	require.NoError(b, err)

	e := jx.GetEncoder()
	// Preallocate internal buffer.
	EncodeMAC(e, mac)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		e.Reset()
		EncodeMAC(e, mac)
	}
}

func BenchmarkDecodeMAC(b *testing.B) {
	data := []byte(`"00:11:22:33:44:55"`)
	d := jx.GetDecoder()

	b.ReportAllocs()
	b.ResetTimer()

	var (
		mac net.HardwareAddr
		err error
	)
	for i := 0; i < b.N; i++ {
		d.ResetBytes(data)
		mac, err = DecodeMAC(d)
	}
	require.NoError(b, err)
	_ = mac
}

func TestMAC(t *testing.T) {
	mac := func(s string) net.HardwareAddr {
		m, err := net.ParseMAC(s)
		require.NoError(t, err)
		return m
	}
	tests := []struct {
		input   string
		wantVal net.HardwareAddr
		wantErr bool
	}{
		// Valid MAC.
		{`"00:11:22:33:44:55"`, mac("00:11:22:33:44:55"), false},
		{`"A1-B2-C3-D4-E5-F6"`, mac("A1-B2-C3-D4-E5-F6"), false},

		// Invalid MAC.
		{"01:23:45:67:89:GH", nil, true}, // GH is not a valid hex digit.
		{`"00-1B-2C-3D-4E"`, nil, true},  // Too few octets.
		{`"A1B2C3D4E5F6"`, nil, true},    // Missing separators.
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), testDecodeEncode(
			DecodeMAC,
			EncodeMAC,
			tt.input,
			tt.wantVal,
			tt.wantErr,
		))
	}
}
