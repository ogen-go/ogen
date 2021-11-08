package json

import (
	"github.com/go-faster/jx"
)

// Decoder is jx.Decoder alias.
type Decoder = jx.Decoder

// GetDecoder gets iterator from pool.
func GetDecoder() *jx.Decoder {
	return jx.GetDecoder()
}

// PutDecoder puts iterator to pool.
func PutDecoder(d *jx.Decoder) {
	jx.PutDecoder(d)
}
