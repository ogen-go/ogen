package json

import (
	"github.com/ogen-go/jx"
)

// Writer is jx.Writer alias.
type Writer = jx.Writer

// GetWriter returns new Writer from pool
func GetWriter() *jx.Writer {
	return jx.GetWriter()
}

// PutWriter puts Writer to pool.
func PutWriter(w *jx.Writer) {
	jx.PutWriter(w)
}
