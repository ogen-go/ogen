package json

import (
	"github.com/ogen-go/jx"
)

// Reader is jx.Reader alias.
type Reader = jx.Reader

// GetReader gets iterator from pool.
func GetReader() *jx.Reader {
	return jx.GetReader()
}

// PutReader puts iterator to pool.
func PutReader(r *jx.Reader) {
	r.Reset(nil)
	r.ResetBytes(nil)
	jx.PutReader(r)
}
