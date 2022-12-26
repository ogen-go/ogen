// Package eventtype defines a custom format for event types.
package eventtype

import (
	"encoding/json"

	"github.com/go-faster/jx"

	"github.com/ogen-go/ogen/gen"
)

type Event = any

// EventFormat defines a custom format for Event.
var EventFormat = gen.CustomFormat[
	Event,
	JSONEventEncoding,
	TextEventEncoding,
]()

// JSONEventEncoding defines a custom JSON encoding for hexadecimal numbers.
type JSONEventEncoding struct{}

// EncodeJSON encodes a hexadecimal number as a JSON string.
func (JSONEventEncoding) EncodeJSON(e *jx.Encoder, v Event) {
	b, err := json.Marshal(v)
	if err != nil {
		e.Null()
		return
	}
	e.Raw(b)
}

// DecodeJSON decodes a hexadecimal number from a JSON string.
func (JSONEventEncoding) DecodeJSON(d *jx.Decoder) (v Event, _ error) {
	r, err := d.Raw()
	if err != nil {
		return v, err
	}
	err = json.Unmarshal(r, &v)
	return v, err
}

// TextEventEncoding defines a custom text encoding for hexadecimal numbers.
type TextEventEncoding struct{}

// EncodeText encodes a hexadecimal number as a string.
func (TextEventEncoding) EncodeText(v Event) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

// DecodeText decodes a hexadecimal number from a string.
func (TextEventEncoding) DecodeText(s string) (v Event, _ error) {
	err := json.Unmarshal([]byte(s), &v)
	return v, err
}
