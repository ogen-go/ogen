package json

import (
	"net/url"
)

func ReadURI(i *Reader) (v url.URL, err error) {
	s, err := i.Str()
	if err != nil {
		return v, err
	}
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return url.URL{}, err
	}
	return *u, nil
}

func WriteURI(s *Writer, v url.URL) {
	s.Str(v.String())
}
