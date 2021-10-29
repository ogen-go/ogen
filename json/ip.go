package json

import (
	"errors"
	"net"
)

func ReadIP(i *Iter) (v net.IP, err error) {
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

func WriteIP(s *Stream, v net.IP) {
	s.WriteString(v.String())
}
