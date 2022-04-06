package json

import (
	"net/netip"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
)

// DecodeIP decodes netip.Addr.
func DecodeIP(i *jx.Decoder) (v netip.Addr, err error) {
	s, err := i.Str()
	if err != nil {
		return v, err
	}
	v, err = netip.ParseAddr(s)
	if err != nil {
		return v, errors.Wrap(err, "bad ip format")
	}
	return v, nil
}

// EncodeIP encodes netip.Addr.
func EncodeIP(s *jx.Encoder, v netip.Addr) {
	b := make([]byte, 64)
	b = v.AppendTo(b[:0])
	s.ByteStr(b)
}
