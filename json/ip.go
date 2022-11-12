package json

import (
	"net/netip"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
)

func decodeIP(d *jx.Decoder, checkVersion func(addr netip.Addr) bool) (v netip.Addr, err error) {
	raw, err := d.Str()
	if err != nil {
		return v, err
	}
	v, err = netip.ParseAddr(raw)
	if err != nil {
		return v, errors.Wrap(err, "bad ip format")
	}
	if checkVersion != nil && !checkVersion(v) {
		return v, errors.New("wrong ip version")
	}
	return v, nil
}

// DecodeIP decodes netip.Addr.
func DecodeIP(d *jx.Decoder) (netip.Addr, error) {
	return decodeIP(d, nil)
}

// DecodeIPv4 decodes netip.Addr.
func DecodeIPv4(d *jx.Decoder) (netip.Addr, error) {
	return decodeIP(d, netip.Addr.Is4)
}

// DecodeIPv6 decodes netip.Addr.
func DecodeIPv6(d *jx.Decoder) (netip.Addr, error) {
	return decodeIP(d, netip.Addr.Is6)
}

// EncodeIP encodes netip.Addr.
func EncodeIP(s *jx.Encoder, v netip.Addr) {
	b := make([]byte, 64)
	b = v.AppendTo(b[:0])
	s.ByteStr(b)
}

// EncodeIPv4 encodes netip.Addr.
func EncodeIPv4(s *jx.Encoder, v netip.Addr) {
	EncodeIP(s, v)
}

// EncodeIPv6 encodes netip.Addr.
func EncodeIPv6(s *jx.Encoder, v netip.Addr) {
	EncodeIP(s, v)
}
