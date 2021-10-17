package json

import (
	"io"

	json "github.com/json-iterator/go"
)

type Stream = json.Stream

func NewStream(w io.Writer) *Stream {
	return json.NewStream(ConfigDefault, w, 1024)
}
