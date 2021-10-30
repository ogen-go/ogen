package json

import (
	"github.com/ogen-go/jx"
)

// Encoder is jx.Encoder alias.
type Encoder = jx.Encoder

// GetEncoder returns new Encoder from pool
func GetEncoder() *jx.Encoder {
	return jx.GetEncoder()
}

// PutEncoder puts Encoder to pool.
func PutEncoder(e *jx.Encoder) {
	jx.PutEncoder(e)
}
