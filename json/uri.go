package json

import (
	"net/url"
)

func ReadURI(i *Decoder) (v url.URL, err error) {
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

func WriteURI(s *Encoder, v url.URL) {
	s.Str(v.String())
}
