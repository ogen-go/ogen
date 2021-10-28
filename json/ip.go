package json

import (
	"errors"
	"net"
)

func ReadIP(i *Iter) (v net.IP, err error) {
	v = net.ParseIP(i.Str())
	if len(v) == 0 {
		return nil, errors.New("bad ip format")
	}
	return v, nil
}

func WriteIP(s *Stream, v net.IP) {
	s.WriteString(v.String())
}
