package json

import (
	"net"
	"strings"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
)

// DecodeMAC decodes net.HardwareAddr.
func DecodeMAC(d *jx.Decoder) (net.HardwareAddr, error) {
	raw, err := d.Str()
	if err != nil {
		return nil, err
	}
	// Keep behavior stable across Go versions by requiring an explicit
	// MAC separator.
	if !strings.ContainsAny(raw, ":-.") {
		return nil, errors.New("invalid MAC address format")
	}
	return net.ParseMAC(raw)
}

// EncodeMAC encodes net.HardwareAddr.
func EncodeMAC(e *jx.Encoder, v net.HardwareAddr) {
	e.Str(v.String())
}
