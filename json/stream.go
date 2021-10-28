package json

import (
	"io"
	"sync"

	"github.com/ogen-go/jir"
)

type Stream = jir.Stream

func NewStream(w io.Writer) *Stream {
	return jir.NewStream(ConfigDefault, w, 1024)
}

func NewCustomStream(cfg API, out io.Writer, bufSize int) *Stream {
	return jir.NewStream(cfg, out, bufSize)
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
