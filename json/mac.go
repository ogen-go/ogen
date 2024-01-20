package json

import (
	"net"

	"github.com/go-faster/jx"
)

// DecodeMAC decodes net.HardwareAddr.
func DecodeMAC(d *jx.Decoder) (net.HardwareAddr, error) {
	raw, err := d.Str()
	if err != nil {
		return nil, err
	}
	return net.ParseMAC(raw)
}

// EncodeMAC encodes net.HardwareAddr.
func EncodeMAC(e *jx.Encoder, v net.HardwareAddr) {
	e.Str(v.String())
}
