// Package json contains helper functions for encoding and decoding JSON.
package json

import (
	std "encoding/json"

	"github.com/go-faster/jx"
)

// Marshal value to json.
func Marshal(val any) ([]byte, error) {
	return std.Marshal(val)
}

// Unmarshal value from json.
func Unmarshal(data []byte, val any) error {
	return std.Unmarshal(data, val)
}

// Unmarshaler implements json reading.
type Unmarshaler interface {
	Decode(d *jx.Decoder) error
}

// Marshaler implements json writing.
type Marshaler interface {
	Encode(e *jx.Encoder)
}

// Value represents a json value.
type Value interface {
	Marshaler
	Unmarshaler
}

// Settable value can be set (present) or unset
// (i.e. not provided or undefined).
type Settable interface {
	IsSet() bool
}

// Resettable value can be unset.
type Resettable interface {
	Reset()
}

// Nullable can be nil (but defined) or not.
type Nullable interface {
	IsNil() bool
}

// Encode Marshaler to byte slice.
func Encode(m Marshaler) []byte {
	e := &jx.Encoder{}
	m.Encode(e)
	return e.Bytes()
}
