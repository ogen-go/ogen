package json

import (
	"github.com/ogen-go/jx"
)

// Iter is jx.Iter alias.
type Iter = jx.Iter

// GetIter gets iterator from pool.
func GetIter() *jx.Iter {
	return jx.Default.GetIter(nil)
}

// PutIter puts iterator to pool.
func PutIter(i *jx.Iter) {
	i.Reset(nil)
	i.ResetBytes(nil)
	jx.Default.PutIter(i)
}
