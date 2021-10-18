package json

import (
	"errors"
	"net"
)

func ReadIP(i *Iterator) (v net.IP, err error) {
	v = net.ParseIP(i.ReadString())
	if len(v) == 0 {
		return nil, errors.New("bad ip format")
	}
	return v, nil
}

func WriteIP(s *Stream, v net.IP) {
	s.WriteString(v.String())
}
