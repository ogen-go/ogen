package json

import (
	"net/url"

	"github.com/go-faster/jx"
)

// DecodeURI decodes url.URL from json.
func DecodeURI(i *jx.Decoder) (v url.URL, err error) {
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

// EncodeURI encodes url.URL to json.
func EncodeURI(s *jx.Encoder, v url.URL) {
	s.Str(v.String())
}
