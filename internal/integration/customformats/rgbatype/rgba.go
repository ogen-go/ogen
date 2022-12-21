// Package rgbatype defines a custom format for RGBA colors.
package rgbatype

import (
	"fmt"

	"github.com/go-faster/jx"
)

type RGBA struct {
	R, G, B, A uint8
}

type JSONRGBAEncoding struct{}

func (JSONRGBAEncoding) EncodeJSON(e *jx.Encoder, v RGBA) {
	var t TextRGBAEncoding
	e.Str(t.EncodeText(v))
}

func (JSONRGBAEncoding) DecodeJSON(d *jx.Decoder) (v RGBA, _ error) {
	s, err := d.Str()
	if err != nil {
		return v, err
	}
	var t TextRGBAEncoding
	return t.DecodeText(s)
}

type TextRGBAEncoding struct{}

func (TextRGBAEncoding) EncodeText(v RGBA) string {
	return fmt.Sprintf("rgba(%d,%d,%d,%d)", v.R, v.G, v.B, v.A)
}

func (TextRGBAEncoding) DecodeText(s string) (v RGBA, _ error) {
	_, err := fmt.Sscanf(s, "rgba(%d,%d,%d,%d)", &v.R, &v.G, &v.B, &v.A)
	return v, err
}
