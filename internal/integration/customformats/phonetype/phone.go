// Package phonetype defines a custom format for phone numbers.
package phonetype

import (
	"regexp"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
)

type Phone string

var phoneRegexp = regexp.MustCompile(`^\+\d+$`)

type JSONPhoneEncoding struct{}

func (JSONPhoneEncoding) EncodeJSON(e *jx.Encoder, v Phone) {
	var t TextPhoneEncoding
	e.Str(t.EncodeText(v))
}

func (JSONPhoneEncoding) DecodeJSON(d *jx.Decoder) (v Phone, _ error) {
	s, err := d.Str()
	if err != nil {
		return v, err
	}
	var t TextPhoneEncoding
	return t.DecodeText(s)
}

type TextPhoneEncoding struct{}

func (TextPhoneEncoding) EncodeText(v Phone) string {
	return string(v)
}

func (TextPhoneEncoding) DecodeText(s string) (v Phone, _ error) {
	if !phoneRegexp.MatchString(s) {
		return v, errors.Errorf("invalid phone %q", s)
	}
	return Phone(s), nil
}
