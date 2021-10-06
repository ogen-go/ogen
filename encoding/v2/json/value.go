package json

import json "github.com/json-iterator/go"

// Unmarshaler implements json reading.
type Unmarshaler interface {
	ReadJSON(i *json.Iterator) bool
}

// Marshaler implements json writing.
type Marshaler interface {
	WriteFieldJSON(k string, s *json.Stream)
	WriteJSON(s *json.Stream)
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

// Nullable can be nil (but defined) or not.
type Nullable interface {
	IsNil() bool
}
