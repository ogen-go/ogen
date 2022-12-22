// Package hextype defines a custom format for hexadecimal numbers.
package hextype

import (
	"strconv"

	"github.com/go-faster/jx"

	"github.com/ogen-go/ogen/gen"
)

// HexFormat defines a custom format for hexadecimal numbers.
var HexFormat = gen.CustomFormat[
	int64,
	JSONHexEncoding,
	TextHexEncoding,
]()

// JSONHexEncoding defines a custom JSON encoding for hexadecimal numbers.
type JSONHexEncoding struct{}

// EncodeJSON encodes a hexadecimal number as a JSON string.
func (JSONHexEncoding) EncodeJSON(e *jx.Encoder, v int64) {
	var t TextHexEncoding
	e.Str(t.EncodeText(v))
}

// DecodeJSON decodes a hexadecimal number from a JSON string.
func (JSONHexEncoding) DecodeJSON(d *jx.Decoder) (v int64, _ error) {
	s, err := d.Str()
	if err != nil {
		return v, err
	}
	var t TextHexEncoding
	return t.DecodeText(s)
}

// TextHexEncoding defines a custom text encoding for hexadecimal numbers.
type TextHexEncoding struct{}

// EncodeText encodes a hexadecimal number as a string.
func (TextHexEncoding) EncodeText(v int64) string {
	return strconv.FormatInt(v, 16)
}

// DecodeText decodes a hexadecimal number from a string.
func (TextHexEncoding) DecodeText(s string) (v int64, _ error) {
	return strconv.ParseInt(s, 16, 64)
}
