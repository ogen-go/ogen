package json

import (
	"net"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
)

// DecodeIP decodes net.IP.
func DecodeIP(i *jx.Decoder) (v net.IP, err error) {
	s, err := i.Str()
	if err != nil {
		return nil, err
	}
	v = net.ParseIP(s)
	if len(v) == 0 {
		return nil, errors.New("bad ip format")
	}
	return v, nil
}

// EncodeIP encodes net.IP.
func EncodeIP(s *jx.Encoder, v net.IP) {
	s.Str(v.String())
}
