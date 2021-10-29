package json

import (
	"io"

	"github.com/ogen-go/jx"
)

// Stream is jx.Stream alias.
type Stream = jx.Stream

// GetStream returns new Stream from pool
func GetStream(w io.Writer) *jx.Stream {
	return jx.Default.GetStream(w)
}

// PutStream puts stream to pool.
func PutStream(s *jx.Stream) {
	jx.Default.PutStream(s)
}
