package json

import (
	"sync"

	"github.com/ogen-go/json"
)

type Iterator = json.Iterator

func NewIterator() *Iterator {
	return json.NewIterator(ConfigDefault)
}

var iterPool = sync.Pool{
	New: func() interface{} {
		return NewIterator()
	},
}

func GetIterator() *Iterator {
	return iterPool.Get().(*Iterator)
}

func PutIterator(i *Iterator) {
	i.Reset(nil)
	i.ResetBytes(nil)
	i.Error = nil
	iterPool.Put(i)
}
