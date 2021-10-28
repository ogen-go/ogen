package json

import (
	"io"
	"sync"

	json "github.com/ogen-go/json"
)

type Stream = json.Stream

func NewStream(w io.Writer) *Stream {
	return json.NewStream(ConfigDefault, w, 1024)
}

func NewCustomStream(cfg API, out io.Writer, bufSize int) *Stream {
	return json.NewStream(cfg, out, bufSize)
}

var streamPool = sync.Pool{
	New: func() interface{} {
		return NewStream(nil)
	},
}

func GetStream(w io.Writer) *Stream {
	s := streamPool.Get().(*Stream)
	s.Reset(w)
	return s
}

func PutStream(s *Stream) {
	s.Reset(nil)
	streamPool.Put(s)
}
