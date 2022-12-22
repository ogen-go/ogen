// Package phonetype defines a custom format for phone numbers.
package phonetype

import (
	"regexp"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"

	"github.com/ogen-go/ogen/gen"
)

// PhoneFormat defines a custom format for phone numbers.
var PhoneFormat = gen.CustomFormat[
	Phone,
	JSONPhoneEncoding,
	TextPhoneEncoding,
]()

// Phone is a phone number.
type Phone string

var phoneRegexp = regexp.MustCompile(`^\+\d+$`)

// JSONPhoneEncoding defines a custom JSON encoding for phone numbers.
type JSONPhoneEncoding struct{}

// EncodeJSON encodes a phone number as a JSON string.
func (JSONPhoneEncoding) EncodeJSON(e *jx.Encoder, v Phone) {
	var t TextPhoneEncoding
	e.Str(t.EncodeText(v))
}

// DecodeJSON decodes a phone number from a JSON string.
func (JSONPhoneEncoding) DecodeJSON(d *jx.Decoder) (v Phone, _ error) {
	s, err := d.Str()
	if err != nil {
		return v, err
	}
	var t TextPhoneEncoding
	return t.DecodeText(s)
}

// TextPhoneEncoding defines a custom text encoding for phone numbers.
type TextPhoneEncoding struct{}

// EncodeText encodes a phone number as a string.
func (TextPhoneEncoding) EncodeText(v Phone) string {
	return string(v)
}

// DecodeText decodes a phone number from a string.
func (TextPhoneEncoding) DecodeText(s string) (v Phone, _ error) {
	if !phoneRegexp.MatchString(s) {
		return v, errors.Errorf("invalid phone %q", s)
	}
	return Phone(s), nil
}
