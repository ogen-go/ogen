package json

import (
	"bytes"
	"sync"
)

var pool = &sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func GetBuffer() *bytes.Buffer {
	return pool.Get().(*bytes.Buffer)
}

func PutBuffer(b *bytes.Buffer) {
	b.Reset()
	pool.Put(b)
}
