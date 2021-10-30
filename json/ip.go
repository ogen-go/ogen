package json

import (
	"errors"
	"net"
)

func ReadIP(i *Decoder) (v net.IP, err error) {
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

func WriteIP(s *Encoder, v net.IP) {
	s.Str(v.String())
}
