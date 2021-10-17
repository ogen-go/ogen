package json

import (
	"io"

	json "github.com/json-iterator/go"
)

type Stream = json.Stream

func NewStream(w io.Writer) *Stream {
	return json.NewStream(ConfigDefault, w, 1024)
}

func NewCustomStream(cfg API, out io.Writer, bufSize int) *Stream {
	return json.NewStream(cfg, out, bufSize)
}
