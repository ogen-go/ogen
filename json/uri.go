package json

import (
	"net/url"
)

func ReadURI(i *Iter) (v url.URL, err error) {
	u, err := url.ParseRequestURI(i.Str())
	if err != nil {
		return url.URL{}, err
	}
	return *u, nil
}

func WriteURI(s *Stream, v url.URL) {
	s.WriteString(v.String())
}
