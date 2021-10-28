package json

import (
	"sync"

	"github.com/ogen-go/jir"
)

// Iter is jir.Iterator alias.
type Iter = jir.Iterator

func newIter() *Iter {
	return jir.NewIterator(ConfigDefault)
}

var iterPool = sync.Pool{
	New: func() interface{} {
		return newIter()
	},
}

// GetIter gets iterator from pool.
func GetIter() *Iter {
	return iterPool.Get().(*Iter)
}

// PutIter puts iterator to pool.
func PutIter(i *Iter) {
	i.Reset(nil)
	i.ResetBytes(nil)
	i.Error = nil
	iterPool.Put(i)
}
