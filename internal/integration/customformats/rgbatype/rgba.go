// Package rgbatype defines a custom format for RGBA colors.
package rgbatype

import (
	"fmt"

	"github.com/go-faster/jx"

	"github.com/ogen-go/ogen/gen"
)

// RGBAFormat defines a custom format for RGBA colors.
var RGBAFormat = gen.CustomFormat[
	RGBA,
	JSONRGBAEncoding,
	TextRGBAEncoding,
]()

// RGBA is a color with red, green, blue, and alpha components.
type RGBA struct {
	R, G, B, A uint8
}

// JSONRGBAEncoding defines a custom JSON encoding for RGBA colors.
type JSONRGBAEncoding struct{}

// EncodeJSON encodes an RGBA color as a JSON string.
func (JSONRGBAEncoding) EncodeJSON(e *jx.Encoder, v RGBA) {
	var t TextRGBAEncoding
	e.Str(t.EncodeText(v))
}

// DecodeJSON decodes an RGBA color from a JSON string.
func (JSONRGBAEncoding) DecodeJSON(d *jx.Decoder) (v RGBA, _ error) {
	s, err := d.Str()
	if err != nil {
		return v, err
	}
	var t TextRGBAEncoding
	return t.DecodeText(s)
}

// TextRGBAEncoding defines a custom text encoding for RGBA colors.
type TextRGBAEncoding struct{}

// EncodeText encodes an RGBA color as a string.
func (TextRGBAEncoding) EncodeText(v RGBA) string {
	return fmt.Sprintf("rgba(%d,%d,%d,%d)", v.R, v.G, v.B, v.A)
}

// DecodeText decodes an RGBA color from a string.
func (TextRGBAEncoding) DecodeText(s string) (v RGBA, _ error) {
	_, err := fmt.Sscanf(s, "rgba(%d,%d,%d,%d)", &v.R, &v.G, &v.B, &v.A)
	return v, err
}
